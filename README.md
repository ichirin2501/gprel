# gprel

[![Build Status](https://github.com/ichirin2501/gprel/workflows/Test/badge.svg?branch=master)](https://github.com/ichirin2501/gprel/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/ichirin2501/gprel)](https://goreportcard.com/report/github.com/ichirin2501/gprel)
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)
[![Coverage Status](https://coveralls.io/repos/github/ichirin2501/gprel/badge.svg?branch=master)](https://coveralls.io/github/ichirin2501/gprel?branch=master)

Golang Purge RElay Logs  
inspired by https://github.com/yoshinorim/mha4mysql-node/blob/master/bin/purge_relay_logs

### Motivation
I'm using [orchestrator](https://github.com/openark/orchestrator) to achieve MySQL HA.  
Pseudo-GTID may not exist in relay-log when orchestrator uses relay-log, so I want set `relay_log_purge=0` and purge relay-log.  
I made this tool to purge the relaylog safely, leaving the Pseudo-GTID.  

## License
MIT

## Author
Motoaki Nishikawa (a.k.a. ichirin2501)
