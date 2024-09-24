<?php

//Observer::$address = 'localhost';
//Observer::$port = 7777;
while (true) {
  MetricObserver::log("database_calls", "file", "file.php@L123", 1 / 7);
  usleep(100 * 1000);
}

class MetricObserver
{
  public static string $address = 'localhost';
  public static int $port = 7777;

  private static ?Socket $socket = null;
  private static bool $connected = false;
  private static int $connectAt = 0;

  public static function log(string $metricName, string $tagName, string $tagValue, float $duration)
  {
    if (!self::$socket) {
      self::$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP) ?: null;
      socket_set_option(self::$socket, SOL_SOCKET, SO_SNDTIMEO, array('sec' => 0, 'usec' => 1));
      self::$connected = false;
    }
    if (!self::$connected) {
      $now = time();
      if (self::$connectAt != $now) {
        self::$connectAt = $now;
        self::$connected = @socket_connect(self::$socket, self::$address, self::$port);
      }
    }
    if (self::$connected) {
      $line = json_encode(["k" => [$metricName, $tagName, $tagValue], "v" => $duration]) . "\n";
      if (!@socket_write(self::$socket, $line, strlen($line))) {
        self::$socket = null;
        self::$connected = false;
      }
    }
  }
}
