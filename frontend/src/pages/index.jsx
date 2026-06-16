import { UserManageCard } from "../components/UserManageCard";
import { StatusCard } from "../components/StatusCard";
import { LinkManageCard } from "../components/LinkManageCard";
import { useGetConfigs } from "../apis/config";
import { Button } from "../components/ui/button";
import { Modal } from "../components/Modal";
import { SettingsCard } from "../components/SettingsCard";
import { LetsEncrypt } from "./start/letsencrypt";
import { useLanguageStore } from "../store/useLanguageStore";
import { Alert } from "../components/Alert";

export default function Index() {
    const { data: config, loaded } = useGetConfigs()
    const { t } = useLanguageStore()
    return (
        <div className='grid grid-cols-1 gap-4'>
            {loaded && !config?.ssl && (
                <Alert
                    type='warning'
                    title={t('ssl_not_configured')}
                    description={t('ssl_warning_index')}
                >
                    <Modal
                        content={(
                            <LetsEncrypt />
                        )}
                        title={t('configure_ssl')}
                    >
                        <Button variant='outline' className='text-white hover:text-white cursor-pointer rounded-full bg-red-500 hover:bg-red-500/80'>{t('configure_ssl')}</Button>
                    </Modal>
                </Alert>
            )}
            {loaded && !config?.has_password && (
                <Alert
                    type='danger'
                    title={t('password_not_set')}
                    description={t('password_warning_desc')}
                >
                    <Modal
                        content={(
                            <SettingsCard />
                        )}
                        title={t('system_settings')}
                    >
                        <Button variant='outline' className='text-white hover:text-white cursor-pointer rounded-full bg-red-500 hover:bg-red-500/80'>{t('set_password')}</Button>
                    </Modal>
                </Alert>
            )}
            <StatusCard />
            <UserManageCard />
            <LinkManageCard />
        </div>
    )
}