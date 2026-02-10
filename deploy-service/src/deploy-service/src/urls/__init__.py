from tornado.web import Application

from .health import urls as urls_health
from .client import urls as urls_client
from .communication import urls as urls_communication
from .container import urls as urls_container

application = Application(handlers=[
    *urls_health,
    *urls_client,
    *urls_communication,
    *urls_container,
])

__all__ = [
    "application"
]
