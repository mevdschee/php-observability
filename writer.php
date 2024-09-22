<?php

while (true) {
  if (Observer::logging()) {
    Observer::log("database_calls", "file", "file.php@L123", 1 / 7);
  }
  usleep(100 * 1000);
}

class Observer
{
  private static ?Socket $socket = null;
  private static bool $connected = false;
  private static int $connectAt = 0;

  public static function logging(bool $connect = true): bool
  {
    if (!self::$socket) {
      self::$socket = socket_create(AF_INET, SOCK_STREAM, SOL_TCP) ?: null;
      self::$connected = false;
    }
    if (!self::$connected) {
      $now = time();
      if ($connect && self::$connectAt != $now) {
        self::$connectAt = $now;
        self::$connected = @socket_connect(self::$socket, 'localhost', '7777');
      }
    }
    return self::$connected;
  }

  public static function log(string $name, string $tagName, string $tag, float $duration)
  {
    if (self::$connected) {
      $line = sprintf("%s:%s:%s:%g", $name, $tagName, $tag, $duration);
      if (!@socket_write(self::$socket, $line . "\n", strlen($line) + 1)) {
        self::$socket = null;
        self::$connected = false;
      }
    }
  }
}
