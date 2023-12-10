# ZXHN-H298A-Manager

This project can update the ipv6 filter to the ZXHN-H298A router
with the current device's ipv6 address.

It also provides the ability to synchronize this IP with the DNS vendor.

## Usage

1. Please set the following environment variables
    1. **ZTE_ADMIN_PSW**: the password
    2. **SERVER_PORT**: which port this server listens
    3. **EXECUTE_INTERVAL**: The interval of each synchronization
2. You can also request the **/execute** to synchronize immediately.