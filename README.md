# php-observability

To run the server:

    go run .

In bash run:

    for run in {1..100}; do php writer.php & done

And to stop:

    killall php

Now observe the metrics:

http://localhost:4000/

NB: The metrics are Prometheus compatible and follow the [OpenMetrics specification](https://github.com/OpenObservability/OpenMetrics/).

Example metrics:

    # HELP database_calls_seconds A summary of the database calls.
    # TYPE database_calls_seconds summary
    database_calls_count{file="file.php@L123"} 1100
    database_calls_sum{file="file.php@L123"} 157.143
    # HELP database_calls_seconds A histogram of the database calls.
    # TYPE database_calls_seconds histogram
    database_calls_seconds_bucket{le="0.001"} 0
    database_calls_seconds_bucket{le="0.01"} 0
    database_calls_seconds_bucket{le="0.1"} 0
    database_calls_seconds_bucket{le="1"} 1100
    database_calls_seconds_bucket{le="10"} 1100
    database_calls_seconds_bucket{le="100"} 1100
    database_calls_seconds_bucket{le="+Inf"} 1100
    database_calls_seconds_sum 157.143
    database_calls_seconds_count 1100

Enjoy!