import pytest

def pytest_addoption(parser):
    parser.addoption(
        "--host",action="store",default="localhost",help="数据库host"
    )
    parser.addoption(
        "--port",action="store",default="3306",help="数据库port"
    )
    parser.addoption(
        "--user",action="store",default="root",help="数据库用户名"
    )
    parser.addoption(
        "--password",action="store",default="",help="数据库密码"
    )
    parser.addoption(
        "--database1",action="store",default="testdb1",help="数据库名称"
    )
    parser.addoption(
        "--database2",action="store",default="testdb2",help="数据库名称"
    )
@pytest.fixture
def host(request):
    return request.config.getoption("--host")

@pytest.fixture
def port(request):
    return request.config.getoption("--port")

@pytest.fixture
def user(request):
    return request.config.getoption("--user")

@pytest.fixture
def password(request):
    return request.config.getoption("--password")

@pytest.fixture
def database1(request):
    return request.config.getoption("--database1")

@pytest.fixture
def database2(request):
    return request.config.getoption("--database2")