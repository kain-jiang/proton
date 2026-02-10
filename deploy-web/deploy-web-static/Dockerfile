# 拉取基础镜像
FROM acr.aishu.cn/as/env-node:22.04.20251021 AS staticBuilder

# 设置镜像标签
LABEL stage=staticBuilder

# 设置工作目录
WORKDIR /build/

# 复制项目文件
COPY . .

# 安装 node_modules
RUN npm install \
    && npm run build

# 拉取基础镜像
FROM acr.aishu.cn/public/nginx:1.28.0-alpine-latest

# 设置主目录
WORKDIR /AnyShare/DeployWebStatic/

# 拷贝静态文件
COPY --from=staticBuilder \
    /build/dist /AnyShare/DeployWebStatic/

# 拷贝nginx配置文件
COPY --from=staticBuilder \
    /build/nginx/default.conf /etc/nginx/conf.d/default.conf

# 暴露端口
EXPOSE 18800

RUN chmod 777 /var/cache/nginx && \
    chmod 777 /var/log/nginx && \
    chmod 777 /var/run/

# 容器启动默认执行命令
CMD ["nginx", "-g", "daemon off;"]