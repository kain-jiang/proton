from src.handlers.health_handler import AliveHandler, ReadyHandler

urls = [
    # 健康检测
    ("/health/alive", AliveHandler),
    ("/health/ready", ReadyHandler),
]