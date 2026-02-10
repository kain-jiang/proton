import os
import unittest
from unittest import mock

from requests import Response
import sys

import requests

sys.path.append("/root/project/DeployService/")

from src.clients.opa import OPAClient


class TestOPAClient(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.mocked_instance = OPAClient()
        cls.user_arr = ["user1", "user2"]
        cls.os_type = "Windows"

    @mock.patch.dict(os.environ, {
        'POLICY_ENGINE_HOST': 'mocked.host',
        'POLICY_ENGINE_PORT': '8080'
    })
    @mock.patch('requests.post')
    def test_download_strategy(self, mock_post):
        # 创建一个Mock对象来模拟Response
        mock_response = mock.MagicMock(spec=requests.Response)
        mock_response.status_code = 200
        # 设置模拟的JSON字符串作为响应体内容
        # mock_response.text = {
        #     "result": {
        #         "result": "accepted",
        #         "mode": 1, 
        #         "remark": ""
        #     }
        # }
        # 这一步在某些情况下是可选的，但有助于使模拟更接近实际响应行为
        mock_response.json.return_value = {
            "result": {
                "result": "accepted",
                "mode": 1, 
                "remark": ""
            }
        }
        # 然后将这个模拟的响应对象设置为requests.post方法返回的结果
        mock_post.return_value = mock_response

        # 调用被测试方法并验证结果
        expected_data = {
            "input": {
                "user": self.user_arr,
                "client": self.os_type
            }
        }
        result = OPAClient.instance().download_strategy(self.user_arr, self.os_type)

        # 验证请求是否正确发送
        mock_post.assert_called_once_with(
            url="http://mocked.host:8080/api/proton-policy-engine/v2/data/client",
            json=expected_data
        )

        # 验证返回的数据
        self.assertEqual(result,  {
            "result": {
                "result": "accepted",
                "mode": 1, 
                "remark": ""
            }
        })

if __name__ == "__main__":
    unittest.main()