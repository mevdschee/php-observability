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
    database_calls_seconds_count{file="file.php@L123"} 2500
    database_calls_seconds_sum{file="file.php@L123"} 357.142
    # HELP database_calls_total_seconds A histogram of the database calls.
    # TYPE database_calls_total_seconds histogram
    database_calls_total_seconds_bucket{le="0.001"} 0
    database_calls_total_seconds_bucket{le="0.01"} 0
    database_calls_total_seconds_bucket{le="0.1"} 0
    database_calls_total_seconds_bucket{le="1"} 2500
    database_calls_total_seconds_bucket{le="10"} 2500
    database_calls_total_seconds_bucket{le="100"} 2500
    database_calls_total_seconds_bucket{le="+Inf"} 2500
    database_calls_total_seconds_sum 357.142
    database_calls_total_seconds_count 2500

Enjoy!