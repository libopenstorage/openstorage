# libopenstorage
libopenstorage is an implementation of the Lib Open Storage specification

### Building:

```bash
# create a 'github.com/libopenstorage' in your GOPATH/src
cd github.com/libopenstorage
git clone https://github.com/libopenstorage/libopenstorage
cd libos
make
sudo make install
```

#### Using libos with systemd

```service
[Unit]
Description=Lib Open Storage

[Service]
CPUQuota=200%
MemoryLimit=1536M
ExecStart=/usr/local/bin/libopenstorage
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

# Contributing

The specification and code is licensed under the Apache 2.0 license found in 
the `LICENSE` file of this repository.  
