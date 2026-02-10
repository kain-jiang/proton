# 拉取基础镜像
FROM acr.aishu.cn/as/env-node:22.04.20251021 AS serviceBuilder

# 设置镜像标签
LABEL stage=serviceBuilder

# 设置工作目录
WORKDIR /build/

# 复制项目文件
COPY . .

# 安装 node_modules
RUN npm install \
    && npm run build

# 拉取基础镜像
FROM acr.aishu.cn/as/env-node:22.04.20251021

# 创建服务目录
RUN mkdir /AnyShare/ \
&& mkdir /AnyShare/DeployWebService \
&& mkdir /AnyShare/tmp \
&& chmod 757 -R /AnyShare/tmp

# 导入前端服务d
COPY --from=serviceBuilder /build/build /AnyShare/DeployWebService/build/
COPY --from=serviceBuilder /build/handlers/views /AnyShare/DeployWebService/build/
COPY --from=serviceBuilder /build/handlers/tools /AnyShare/DeployWebService/build/

# node16忽略https证书
ENV NODE_TLS_REJECT_UNAUTHORIZED=0

# 容器启动默认执行命令
CMD ["node", "/AnyShare/DeployWebService/build/bundle.js"]