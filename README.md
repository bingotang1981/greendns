# greendns
It is a simple DNS server which is based on github.com/miekg/dns. For DNS questions, it will forward them to a user defined DNS server (e.g. 8.8.8.8, 114.114.114.114, etc). For some specified domains, it can reply with some predefined "wrong" ips. You may use the command as follows:
greendns <filterFilePath> [port] [upstreamDns] [defaultFakeIp]. 
