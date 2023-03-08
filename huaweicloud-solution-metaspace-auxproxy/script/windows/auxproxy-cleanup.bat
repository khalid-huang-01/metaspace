@echo off
setlocal

set logfile=C:\"Program Files"\auxproxy\bats.log

echo "[%date:~0,10% %time%][auxproxy-cleanup] find auxproxy.exe" >>%logfile%
tasklist | find /i "auxproxy.exe" || exit

echo "[%date:~0,10% %time%][auxproxy-cleanup] start cleanup" >>%logfile%

set shutdown_code=000
:cleanup_step
if %shutdown_code% neq 202 (
	for /f %%i in ('C:\"Program Files"\auxproxy\curl.exe -X POST -s -w "%%{http_code}" http://localhost:60001/v1/cleanup') do ( set shutdown_code=%%i )
	echo "[%date:~0,10% %time%][auxproxy-cleanup] shutdown_code: %shutdown_code%" >>%logfile%
	ping -n 5 127.0.0.1 > nul
	goto :cleanup_step
)


set shutdown_state="RESERVED"
:shutdown_step
if %shutdown_state% neq "FINISHED" (
	for /f %%i in ('C:\"Program Files"\auxproxy\curl.exe http://localhost:60001/v1/cleanup-state ^| "C:\Program Files\auxproxy\jq.exe" .state') do ( set shutdown_state=%%i )
	echo "[%date:~0,10% %time%][auxproxy-cleanup] shutdown_code: %shutdown_state%" >>%logfile%
	ping -n 5 127.0.0.1 > nul
	goto :shutdown_step
)

echo "[%date:~0,10% %time%][auxproxy-cleanup] finish cleanup" >>%logfile%
