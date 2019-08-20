## gprel
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat)](LICENSE)

Golang Purge RElay Logs  
inspired by https://github.com/yoshinorim/mha4mysql-node/blob/master/bin/purge_relay_logs

### Motivation
I want to use orchestrator's Pseudo-GTID.  
Pseudo-GTID may not exist in relay-log when orchestrator uses relay-log, so I considered set `relay_log_purge=0` and purge relay-log.

Consider the following points
- Pseudo-GTID is included in relay-log even if rotate
- Don't rotate when orchestrator executes `stop slave`
  - e.g. failover etc

## License
MIT

## Author
Motoaki Nishikawa (a.k.a. ichirin2501)
