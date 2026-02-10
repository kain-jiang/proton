from src.clients.config import ConfigClient
from src.clients.protoncli import ProtonCliClient
from src.clients.cms import CMSClient, CMSObject

def init_anyshare_cms():
    anyshare_cms = CMSClient.instance().head_cms_data("anyshare")
    if ConfigClient.load_config().use_protoncli():
        if anyshare_cms is None:
            anyshare_cms = CMSObject.create("anyshare", ConfigClient.load_config().init_access_address())
        anyshare_cms.real_data.update({
            "mode": ProtonCliClient.instance().deploy_mode(),
            "devicespec.conf": ProtonCliClient.instance().devicespce_ini()
        })
        anyshare_cms.save(CMSClient.instance())
    else:
        if anyshare_cms is None:
            raise Exception("anyshare cms not found, you must init this cms when use_protoncli is false")


