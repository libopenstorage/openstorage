# openstorage
openstorage is an implementation of the Open Storage specification

### Building:

```bash
# create a 'github.com/openstorage' in your GOPATH/src
git clone https://github.com/libopenstorage/openstorage
cd openstorage
make
sudo make install
```

#### Using openstorage with systemd

```service
[Unit]
Description=Open Storage

[Service]
CPUQuota=200%
MemoryLimit=1536M
ExecStart=/usr/local/bin/ost
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

# Contributing

The specification and code is licensed under the Apache 2.0 license found in 
the `LICENSE` file of this repository.  
