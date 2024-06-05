# curvelog-vss
VISS Client that plots a time based vss signal and its curved logged data points.

In the settings.json we specify VISS server adress and port.
The signal we want to plot, the time based parameter in milliseconds,
The curve logging maximum error and buffer size.

Below is an example where the VISS server is running locally and set
to listen to gRPC clients at 8887.

```
{
  "vss-name": "vehicle.speed",
  "sub-period-ms": "100",
  "curve-log-err": "1.0",
  "curve-log-buf": "20",
  "adress": "0.0.0.0",
  "port": 8887
}
```

To run this a **VISS** server installtion with a **data feeder** is needed.
