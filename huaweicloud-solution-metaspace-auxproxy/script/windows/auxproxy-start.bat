@echo off

set logfile=C:\"Program Files"\auxproxy\bats.log

echo "[%date:~0,10% %time%][auxproxy-start] sleep 60s" >>%logfile%
ping -n 60 127.0.0.1>nul

echo "[%date:~0,10% %time%][auxproxy-start] auto restart" >>%logfile%
title auto_restart.bat
:check_restart
tasklist | find "auxproxy.exe" || "C:\\Program Files\\auxproxy\\auxproxy.exe" -auxproxy-address 0.0.0.0:60001 -grpc-address 0.0.0.0:60002 -cloud-platform-address http://169.254.169.254
ping -n 30 127.0.0.1>nul
goto check_restart
