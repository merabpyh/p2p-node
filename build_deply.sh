#!/bin/bash

clear && go build . 

scp ./p2p-node root@192.168.1.22:/root/goprogdir/p2p-prog/p2p-node && echo "Cert1 done!"
scp ./p2p-node root@192.168.1.23:/root/goprogdir/p2p-prog/p2p-node && echo "Cert2 done!"

