go run ./cmd/seedsearch -from 1 -to 75000 -runs 120 -top 15 -energie-karte eroeffnungsfeier -finanz-karte schwerindustrie -schichten 3 -schicht-keep 7 -month-filter 1 -spill-memory-mb 4000


go run ./cmd/sim -seed 511 -runs 200 -demand-i 0 -demand-r 1 -demand-w 0 -demand-b 1 -finanz-karte schwerindustrie -month-filter 1 -trace-loop 3 -cost-p1 3 -cost-p2 4