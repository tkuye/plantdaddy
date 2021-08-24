import ujson
import utime

def get_configs():
    """Gets the configurations for the module.
    """
    with open("config.json") as f:
       return ujson.load(f)


def do_connect(ssid: str, password: str) -> bool:
    """
    Connects esp32 to wifi

    Args:
        ssid (str): wifi ssid
        password (str): wifi password
    """
    import network
    sta_if = network.WLAN(network.STA_IF)
    if not sta_if.isconnected():
        print('connecting to network...')
        sta_if.active(True)
        if ssid is None or password is None:
            print("Cannot connect no ssid or password")
            return False
        sta_if.connect(ssid, password)
        time = utime.time()
        while not sta_if.isconnected():
            if time.time() - time > 15:
                return False
    print('network config:', sta_if.ifconfig())
    return True

