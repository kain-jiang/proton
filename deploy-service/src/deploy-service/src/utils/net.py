from IPy import IP


def get_host_for_url(host: str) -> str:
    try:
        version = IP(host).version()
        return f"[{host}]" if version == 6 else host
    except Exception:  # noqa
        return host
