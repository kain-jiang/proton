#仓库: DeployService

## 开发指南: 
1. 创建虚拟环境: `python -m venv .venv`
2. 激活虚拟环境：`.\.venv\Scripts\Activate.ps1` 或者 `source ./.venv/bin/activate`
3. 安装pypi开发依赖：`pip install -r requirements-dev.txt`
> 需要从devops拉取代码，请配置好和devops的ssh认证
4. 安装devops依赖：`python -B deps/install-dev.py`
