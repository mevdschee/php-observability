<?php

class MetricObserver
{
  public static $port = 7777;

  private static $socket = null;
  private static $connected = false;
  private static $connectAt = 0;

  public static function isConnected()
  {
    if (self::$socket === null) {
      self::$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP);
      if (!self::$socket) {
        self::$socket = null;
      }
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

  public static function log($metricName, $labelName, $labelValue, $duration = null)
  {
    if (self::isConnected()) {
      if ($duration === null) {
        $line = json_encode(array($metricName, $labelName, $labelValue));
      } else {
        $line = json_encode(array($metricName, $labelName, $labelValue, (string)$duration));
      }
      if (!@socket_write(self::$socket, $line . "\n", strlen($line) + 1)) {
        self::$socket = null;
        self::$connected = false;
      }
    }
  }
}
