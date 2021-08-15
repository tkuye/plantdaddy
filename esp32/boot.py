# This file is executed on every boot (including wake-boot from deepsleep)
#import esp
#esp.osdebug(None)
#import webrepl
#webrepl.start()
import ujson

def get_configs():
    """Gets the configurations for the module.
    """
    with open("config.json") as f:
       return ujson.load(f)

def do_connect(ssid: str, password: str):
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
            return None
        sta_if.connect(ssid, password)
        while not sta_if.isconnected():
            pass
    print('network config:', sta_if.ifconfig())


if __name__ == "__main__":
    config = get_configs()
    if config is not None:
        do_connect(config.get("ssid"), config.get("password"))
