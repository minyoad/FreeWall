package services

import (
	"context"
	"log"

	"github.com/xtls/xray-core/common/buf"
	"github.com/xtls/xray-core/common/net"
	"github.com/xtls/xray-core/common/session"
	"github.com/xtls/xray-core/features/routing"
	"github.com/xtls/xray-core/transport"
	"golang.org/x/time/rate"
)

type XrayDispatcher struct {
	routing.Dispatcher
	tracker *StatisticsTracker
}

func NewXrayDispatcher(d routing.Dispatcher, tracker *StatisticsTracker) *XrayDispatcher {
	return &XrayDispatcher{
		Dispatcher: d,
		tracker:    tracker,
	}
}

func (d *XrayDispatcher) Dispatch(ctx context.Context, dest net.Destination) (*transport.Link, error) {
	log.Printf("[XrayDispatcher] Dispatch called for dest: %s", dest.String())
	link, err := d.Dispatcher.Dispatch(ctx, dest)
	if err != nil {
		return nil, err
	}

	var email string
	// InboundSourceObject is *protocol.User in some contexts?
	// According to Xray code:
	// if u, ok := content.Attributes["InboundUser"]; ok { ... }
	// Or content.InboundUser but that failed.
	// Let's use what search said: InboundFromContext

	inbound := session.InboundFromContext(ctx)
	if inbound == nil {
		if content := session.ContentFromContext(ctx); content == nil {
			log.Printf("[XrayDispatcher] ContentFromContext is nil")
		} else {
			keys := []string{}
			for k := range content.Attributes {
				keys = append(keys, k)
			}
			log.Printf("[XrayDispatcher] InboundFromContext is nil. Content Attributes: %v", keys)
		}
	} else if inbound.User == nil {
		log.Printf("[XrayDispatcher] Inbound.User is nil. Source: %v, Tag: %s", inbound.Source, inbound.Tag)
	} else {
		email = inbound.User.Email
		log.Printf("[XrayDispatcher] Found user: %s", email)
	}

	// FALLBACK DEBUGGING: Log if missing, but do not default to "unknown" yet
	if email == "" {
		// Check fallback attributes
		if content := session.ContentFromContext(ctx); content != nil {
			if uVal, ok := content.Attributes["InboundUser"]; ok {
				log.Printf("[XrayDispatcher] Found InboundUser in attributes: %T", uVal)
			}
		}
	} else {
		log.Printf("[XrayDispatcher] Found user: %s", email)
	}

	if email == "" {
		return link, nil
	}

	limiter := d.tracker.GetLimiterForUser(email)

	if limiter != nil {
		log.Printf("[XrayDispatcher] Rate limiting user: %s, limit: %.2f", email, limiter.Limit())
	}

	log.Printf("[XrayDispatcher] Wrapping connection for user: %s", email)

	// We need to construct a new Link that wraps the Reader/Writer
	newLink := &transport.Link{
		Reader: link.Reader,
		Writer: link.Writer,
	}

	// Wrap Reader (Downlink? Upstream to client?)
	// Link.Reader is what we read from upstream (remote). Writing to client.
	// Actually, Dispatch returns a Link to write to outbound and read from outbound.
	// So Link.Writer is writing to outbound (Uplink). Link.Reader is reading from outbound (Downlink).

	if link.Reader != nil {
		newLink.Reader = &RateLimitedReader{
			Reader:  link.Reader,
			limiter: limiter,
			ctx:     ctx,
		}
	}

	if link.Writer != nil {
		newLink.Writer = &RateLimitedWriter{
			Writer:  link.Writer,
			limiter: limiter,
			ctx:     ctx,
		}
	}

	return newLink, nil
}

func (d *XrayDispatcher) DispatchLink(ctx context.Context, dest net.Destination, link *transport.Link) error {
	log.Printf("[XrayDispatcher] DispatchLink called for dest: %s", dest.String())

	// Identify user
	var email string
	inbound := session.InboundFromContext(ctx)
	if inbound != nil && inbound.User != nil {
		email = inbound.User.Email
		log.Printf("[XrayDispatcher] DispatchLink Found user: %s", email)
	} else if content := session.ContentFromContext(ctx); content != nil {
		// Fallback check
		if uVal, ok := content.Attributes["InboundUser"]; ok {
			log.Printf("[XrayDispatcher] DispatchLink Found InboundUser in attributes: %T", uVal)
		}
	}

	if email != "" {
		limiter := d.tracker.GetLimiterForUser(email)

		log.Printf("[XrayDispatcher] Wrapping connection in DispatchLink for user: %s", email)

		if link.Reader != nil {
			link.Reader = &RateLimitedReader{
				Reader:  link.Reader,
				limiter: limiter,
				ctx:     ctx,
			}
		}
		if link.Writer != nil {
			link.Writer = &RateLimitedWriter{
				Writer:  link.Writer,
				limiter: limiter,
				ctx:     ctx,
			}
		}
	}

	return d.Dispatcher.DispatchLink(ctx, dest, link)
}

// Xray 1.8+ uses Type(), older used something else.
func (d *XrayDispatcher) Type() interface{} {
	return routing.DispatcherType()
}

func (d *XrayDispatcher) Start() error {
	log.Printf("[XrayDispatcher] Start called")
	return d.Dispatcher.Start()
}

func (d *XrayDispatcher) Close() error {
	log.Printf("[XrayDispatcher] Close called")
	return d.Dispatcher.Close()
}

type RateLimitedWriter struct {
	buf.Writer
	limiter *rate.Limiter
	// ctx is the connection-level context (from Dispatch/DispatchLink).
	// When the connection is torn down, ctx is cancelled, which unblocks
	// any goroutine waiting in WaitN and prevents goroutine leaks.
	ctx context.Context
}

func (w *RateLimitedWriter) WriteMultiBuffer(mb buf.MultiBuffer) error {
	l := int64(mb.Len())
	if l > 0 && w.limiter != nil {
		// Split wait into burst-sized chunks to avoid WaitN(n > burst) errors.
		burst := w.limiter.Burst()
		remaining := int(l)
		for remaining > 0 {
			waitN := remaining
			if waitN > burst {
				waitN = burst
			}
			if err := w.limiter.WaitN(w.ctx, waitN); err != nil {
				// Context cancelled (connection closed) — stop throttling.
				break
			}
			remaining -= waitN
		}
	}
	return w.Writer.WriteMultiBuffer(mb)
}

type RateLimitedReader struct {
	buf.Reader
	limiter *rate.Limiter
	// ctx mirrors the purpose of RateLimitedWriter.ctx.
	ctx context.Context
}

func (r *RateLimitedReader) ReadMultiBuffer() (buf.MultiBuffer, error) {
	mb, err := r.Reader.ReadMultiBuffer()
	if !mb.IsEmpty() && r.limiter != nil {
		burst := r.limiter.Burst()
		remaining := int(int64(mb.Len()))
		for remaining > 0 {
			waitN := remaining
			if waitN > burst {
				waitN = burst
			}
			if werr := r.limiter.WaitN(r.ctx, waitN); werr != nil {
				// Context cancelled — return data already read.
				break
			}
			remaining -= waitN
		}
	}
	return mb, err
}
