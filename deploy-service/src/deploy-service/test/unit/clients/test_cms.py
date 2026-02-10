from src.clients.cms import CMSClient
import unittest

class TestAsyncCms(unittest.TestCase):


    async def async_get_cms_data(self):
        """
        获取cms配置async调用测试用例
        需要在本地搭建cms服务端或调整配置为外部cms客户端
        TODO: 该用例也许不该在这里，后续视乎情况移动
        """
        cli = CMSClient(host="127.0.0.1", port=8080)

        obj = await cli.async_get_cms_data("anyshare")
        print(obj.real_data)
        # assert(obj.real_data, None)
    
    def test_async_get_cms_data(self):
        import asyncio
        asyncio.run(self.async_get_cms_data())

if __name__ == "__main__":
    unittest.main()