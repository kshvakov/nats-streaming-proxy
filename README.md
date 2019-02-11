# nats-streaming-proxy
Write to the [NATS Streaming](https://nats.io/documentation/streaming/nats-streaming-intro/) via Memcached protocol.

### PHP example

```php
<?php
$mem = new Memcached('nats-streaming-pool');
if (count($mem->getServerList()) == 0) {
    $mem->addServer("10.112.179.191", 11211);
    $mem->addServer("10.112.179.192", 11211);
    // http://php.net/manual/en/memcached.constants.php
    $mem->setOption(Memcached::OPT_TCP_NODELAY, true);
    $mem->setOption(Memcached::OPT_COMPRESSION, false);
}
$mem->set('subject', json_encode([
    'event_time' => time(),
    'event_type' => 'type',
    'payload'    => 'XXXX'
]));
```