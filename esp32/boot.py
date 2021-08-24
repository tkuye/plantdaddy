# This file is executed on every boot (including wake-boot from deepsleep)
#import esp
#esp.osdebug(None)
#import webrepl
#webrepl.start()
import pbluetooth
from helpers import get_configs, do_connect
import main



if __name__ == "__main__":
    config = get_configs()
    if config is not None:
        if "ssid" not in config or "password" not in config:
            pbluetooth.main()
        else:
            c = do_connect(config.get("ssid"), config.get("password"))
            if c:
                main.main()
        