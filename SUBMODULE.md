# Submodule 部署指南

本指南说明如何将 CLIProxyAPI 作为 Git submodule 集成到母项目中进行部署。

## 在母项目中添加 Submodule

```bash
# 在母项目根目录执行
git submodule add https://github.com/router-for-me/CLIProxyAPI.git services/cliproxy
git submodule update --init --recursive
```

## 母项目目录结构示例

```
parent-project/
├── docker-compose.yml          # 母项目的 compose 文件
├── .env                        # 母项目环境变量
├── services/
│   ├── cliproxy/              # CLIProxyAPI submodule
│   │   ├── docker-compose.yml
│   │   ├── config.example.yaml
│   │   └── ...
│   ├── other-service/
│   └── ...
└── configs/
    └── cliproxy/
        ├── config.yaml        # 实际配置文件
        └── .env              # CLIProxyAPI 环境变量
```

## 母项目 docker-compose.yml 配置

### 方式一：引用 submodule 的 compose 文件

```yaml
# parent-project/docker-compose.yml
services:
  # 其他服务...
  
  # 引入 CLIProxyAPI
  cli-proxy-api:
    extends:
      file: ./services/cliproxy/docker-compose.yml
      service: cli-proxy-api
    environment:
      # 覆盖或添加环境变量
      DEPLOY: production
    volumes:
      # 使用母项目的配置目录
      - ./configs/cliproxy/config.yaml:/CLIProxyAPI/config.yaml
      - ./configs/cliproxy/auths:/root/.cli-proxy-api
      - ./logs/cliproxy:/CLIProxyAPI/logs
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

### 方式二：直接定义服务

```yaml
# parent-project/docker-compose.yml
services:
  cli-proxy-api:
    image: ${CLI_PROXY_IMAGE:-ghcr.io/leevianchang/cliproxyapi:latest}
    pull_policy: always
    container_name: cli-proxy-api
    env_file:
      - ./configs/cliproxy/.env
    ports:
      - "8317:8317"
    volumes:
      - ./configs/cliproxy/config.yaml:/CLIProxyAPI/config.yaml
      - ./configs/cliproxy/auths:/root/.cli-proxy-api
      - ./logs/cliproxy:/CLIProxyAPI/logs
    restart: unless-stopped
    networks:
      - app-network
    depends_on:
      - postgres  # 如果使用 PostgreSQL
      
  # 其他服务...
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: cliproxy_production
      POSTGRES_USER: cliproxy_user
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - app-network

networks:
  app-network:
    driver: bridge

volumes:
  postgres_data:
```

## 配置文件准备

### 1. 创建配置目录

```bash
mkdir -p configs/cliproxy
mkdir -p logs/cliproxy
```

### 2. 复制配置模板

```bash
cp services/cliproxy/config.example.yaml configs/cliproxy/config.yaml
```

### 3. 编辑配置文件

```bash
vim configs/cliproxy/config.yaml
```

根据需要修改：
- `host` 和 `port`
- `api-keys`
- API 密钥配置（gemini-api-key、claude-api-key 等）
- 存储后端配置

### 4. 创建环境变量文件（可选）

```bash
cat > configs/cliproxy/.env << 'EOF'
# 管理后台密码
MANAGEMENT_PASSWORD=your-secure-password

# PostgreSQL 配置（如果使用）
PGSTORE_DSN=postgresql://cliproxy_user:password@postgres:5432/cliproxy_production

# 其他环境变量...
EOF
```

## 部署流程

### 1. 克隆母项目（包含 submodules）

```bash
git clone --recursive https://github.com/your-org/parent-project.git
cd parent-project
```

如果已经克隆但没有 submodules：
```bash
git submodule update --init --recursive
```

### 2. 准备配置文件

```bash
# 复制并编辑配置
cp services/cliproxy/config.example.yaml configs/cliproxy/config.yaml
vim configs/cliproxy/config.yaml
```

### 3. 启动服务

```bash
# 启动所有服务
docker-compose up -d

# 仅启动 CLIProxyAPI
docker-compose up -d cli-proxy-api

# 查看日志
docker-compose logs -f cli-proxy-api
```

## 更新 Submodule

### 更新到最新版本

```bash
cd services/cliproxy
git pull origin main
cd ../..
git add services/cliproxy
git commit -m "Update CLIProxyAPI submodule"
```

### 更新到特定版本

```bash
cd services/cliproxy
git fetch --tags
git checkout v1.0.0
cd ../..
git add services/cliproxy
git commit -m "Update CLIProxyAPI to v1.0.0"
```

### 更新所有 submodules

```bash
git submodule update --remote --merge
```

## 环境变量配置

### 母项目 .env 文件

```bash
# parent-project/.env

# CLIProxyAPI 镜像版本
CLI_PROXY_IMAGE=ghcr.io/leevianchang/cliproxyapi:latest

# PostgreSQL 配置
POSTGRES_PASSWORD=secure_password_here

# 其他服务配置...
```

### CLIProxyAPI 专用 .env

```bash
# configs/cliproxy/.env

# 管理后台配置
MANAGEMENT_PASSWORD=change-me-to-a-strong-password
MANAGEMENT_SECRET_KEY=mgmt_sk_your_secret_key_here

# PostgreSQL Token Store（可选）
PGSTORE_DSN=postgresql://cliproxy_user:${POSTGRES_PASSWORD}@postgres:5432/cliproxy_production
PGSTORE_SCHEMA=public
PGSTORE_LOCAL_PATH=/var/lib/cliproxy/pgstore

# Git 配置存储（可选）
# GITSTORE_GIT_URL=https://github.com/your-org/cliproxy-config.git
# GITSTORE_GIT_TOKEN=ghp_your_token_here

# S3 对象存储（可选）
# OBJECTSTORE_ENDPOINT=https://s3.amazonaws.com
# OBJECTSTORE_BUCKET=cliproxy-config
# OBJECTSTORE_ACCESS_KEY=your_access_key
# OBJECTSTORE_SECRET_KEY=your_secret_key
```

## 网络配置

如果 CLIProxyAPI 需要与其他服务通信：

```yaml
services:
  cli-proxy-api:
    # ...
    networks:
      - app-network
      - external-network
    
  your-app:
    # ...
    environment:
      # 使用服务名访问 CLIProxyAPI
      CLIPROXY_URL: http://cli-proxy-api:8317
    networks:
      - app-network

networks:
  app-network:
    internal: true  # 内部网络
  external-network:
    # 外部访问
```

## 反向代理配置（Nginx 示例）

```nginx
# nginx.conf
upstream cliproxy {
    server cli-proxy-api:8317;
}

server {
    listen 80;
    server_name api.example.com;
    
    location / {
        proxy_pass http://cliproxy;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # WebSocket 支持
        proxy_read_timeout 86400;
    }
}
```

## 持久化数据

确保重要数据持久化：

```yaml
volumes:
  # CLIProxyAPI 配置和认证数据
  - ./configs/cliproxy/config.yaml:/CLIProxyAPI/config.yaml:ro
  - ./configs/cliproxy/auths:/root/.cli-proxy-api
  - ./logs/cliproxy:/CLIProxyAPI/logs
  
  # 如果使用文件存储后端
  - cliproxy_pgstore:/var/lib/cliproxy/pgstore
  - cliproxy_gitstore:/data/cliproxy/gitstore
  - cliproxy_objectstore:/data/cliproxy/objectstore

volumes:
  cliproxy_pgstore:
  cliproxy_gitstore:
  cliproxy_objectstore:
```

## 健康检查

添加健康检查确保服务正常运行：

```yaml
services:
  cli-proxy-api:
    # ...
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:8317/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
```

## 故障排查

### 查看日志

```bash
# 查看所有服务日志
docker-compose logs -f

# 仅查看 CLIProxyAPI 日志
docker-compose logs -f cli-proxy-api

# 查看最近 100 行
docker-compose logs --tail=100 cli-proxy-api
```

### 进入容器调试

```bash
docker-compose exec cli-proxy-api sh
```

### 检查配置

```bash
# 验证配置文件语法
docker-compose config

# 查看实际使用的配置
docker-compose exec cli-proxy-api cat /CLIProxyAPI/config.yaml
```

### 重启服务

```bash
# 重启 CLIProxyAPI
docker-compose restart cli-proxy-api

# 重新构建并启动
docker-compose up -d --force-recreate cli-proxy-api
```

## 安全建议

1. **配置文件权限**
   ```bash
   chmod 600 configs/cliproxy/config.yaml
   chmod 600 configs/cliproxy/.env
   ```

2. **不要提交敏感信息**
   ```bash
   # 母项目 .gitignore
   configs/cliproxy/config.yaml
   configs/cliproxy/.env
   configs/cliproxy/auths/
   logs/
   ```

3. **使用环境变量**
   - 敏感信息通过环境变量传递
   - 使用 Docker secrets（Swarm 模式）

4. **网络隔离**
   - 使用内部网络隔离服务
   - 仅暴露必要的端口

5. **定期更新**
   ```bash
   # 更新镜像
   docker-compose pull
   docker-compose up -d
   ```

## CI/CD 集成

### GitHub Actions 示例

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          submodules: recursive
          
      - name: Deploy to Server
        uses: appleboy/ssh-action@v1.0.0
        with:
          host: ${{ secrets.SERVER_HOST }}
          username: ${{ secrets.SERVER_USER }}
          key: ${{ secrets.SERVER_SSH_KEY }}
          script: |
            cd ~/parent-project
            git pull --recurse-submodules
            docker-compose pull
            docker-compose up -d
```

## 备份和恢复

### 备份

```bash
#!/bin/bash
# backup.sh
BACKUP_DIR="backups/$(date +%Y%m%d_%H%M%S)"
mkdir -p "$BACKUP_DIR"

# 备份配置
cp -r configs/cliproxy "$BACKUP_DIR/"

# 备份数据库（如果使用 PostgreSQL）
docker-compose exec -T postgres pg_dump -U cliproxy_user cliproxy_production > "$BACKUP_DIR/database.sql"

# 打包
tar -czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
rm -rf "$BACKUP_DIR"
```

### 恢复

```bash
#!/bin/bash
# restore.sh
BACKUP_FILE=$1

# 解压
tar -xzf "$BACKUP_FILE"
BACKUP_DIR="${BACKUP_FILE%.tar.gz}"

# 恢复配置
cp -r "$BACKUP_DIR/cliproxy/"* configs/cliproxy/

# 恢复数据库
docker-compose exec -T postgres psql -U cliproxy_user cliproxy_production < "$BACKUP_DIR/database.sql"
```

## 参考资料

- [CLIProxyAPI 主仓库](https://github.com/router-for-me/CLIProxyAPI)
- [部署指南](./DEPLOYMENT.md)
- [配置示例](./config.example.yaml)
- [Docker Compose 文档](https://docs.docker.com/compose/)
- [Git Submodules 文档](https://git-scm.com/book/en/v2/Git-Tools-Submodules)
