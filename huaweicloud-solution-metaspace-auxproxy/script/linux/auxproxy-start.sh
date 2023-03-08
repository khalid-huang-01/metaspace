#!/bin/bash

/etc/auxproxy/auxproxy -auxproxy-address 0.0.0.0:60001 -grpc-address 0.0.0.0:60002 -cloud-platform-address http://169.254.169.254

