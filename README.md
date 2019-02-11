# nats-streaming-proxy
Write to the NATS Streaming via Memcached protocol

Example

```php
<?php
$mem = new Memcached('nats-streaming-pool');
$mem->addServer('10.112.179.191', 11211);
$mem->set('subject', json_encode([
    'event_time' => time(),
    'event_type' => 'type',
    'payload'    => 'XXXX'
]));
```