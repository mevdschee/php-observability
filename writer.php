<?php

class MetricObserver
{
  public static int $port = 7777;

  private static ?Socket $socket = null;
  private static bool $connected = false;
  private static int $connectAt = 0;

  public static function isConnected(): bool
  {
    if (self::$socket === null) {
      self::$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP) ?: null;
      self::$connected = false;
    }
    if (!self::$connected) {
      $now = time();
      if (self::$connectAt != $now) {
        self::$connectAt = $now;
        self::$connected = @socket_connect(self::$socket, 'localhost', self::$port);
      }
    }
    return self::$connected;
  }

  public static function log(string $metricName, string $labelName, string $labelValue, ?float $duration = null)
  {
    if (self::isConnected()) {
      if ($duration === null) {
        $line = json_encode([$metricName, $labelName, $labelValue]);
      } else {
        $line = json_encode([$metricName, $labelName, $labelValue, (string)$duration]);
      }
      if (!@socket_write(self::$socket, $line . "\n", strlen($line) + 1)) {
        self::$socket = null;
        self::$connected = false;
      }
    }
  }
}
