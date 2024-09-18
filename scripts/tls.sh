cfssl selfsign -config cfssl.json --profile rootca "Teleport CA" csr.json | cfssljson -bare root

cfssl genkey csr.json | cfssljson -bare server
cfssl genkey csr.json | cfssljson -bare client

cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile server server.csr | cfssljson -bare server
cfssl sign -ca root.pem -ca-key root-key.pem -config cfssl.json -profile client client.csr | cfssljson -bare client