<?php

include 'MetricObserver.php';

while (true) {
  MetricObserver::log("database_calls", "file", "src/Controller/UserController.php@L123", 0.142857);
  usleep(100 * 1000);
}
