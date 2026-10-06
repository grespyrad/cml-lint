#!/usr/bin/env python3
"""Повторные rotated Go benchmark tests стадий с сохранением сырых строк."""
import argparse
import json
import os
import re
import statistics
import subprocess
from pathlib import Path

root=Path(__file__).resolve().parents[1]
parser=argparse.ArgumentParser()
parser.add_argument('--rounds',type=int,default=5)
parser.add_argument('--benchtime',default='300ms')
args=parser.parse_args()
env={**os.environ,'GOWORK':'off','GOMAXPROCS':'1'}
module=root/'bench/lex'
subprocess.run(['go','test','./...'],cwd=module,env=env,check=True)
cases=[f'{model}/{stage}' for model in ['service-before','service'] for stage in ['Tokenize/CMLGo','Tokenize/Participle','CheckSyntax','Parse','Validate','Unmarshal']]
rows=[]
raw=[]
for round_number in range(args.rounds):
    order=cases[round_number:]+cases[:round_number]
    for case in order:
        pattern='/'.join('^'+re.escape(part)+'$' for part in ['BenchmarkStages',*case.split('/')])
        command=['go','test','-run=^$','-bench='+pattern,'-benchmem','-benchtime='+args.benchtime,'-count=1']
        output=subprocess.check_output(command,cwd=module,env=env,text=True)
        raw.append(output)
        match=re.search(r'BenchmarkStages/(\S+)\s+\d+\s+([\d.]+) ns/op\s+[\d.]+ MB/s\s+(\d+) B/op\s+(\d+) allocs/op',output)
        if not match:raise RuntimeError('missing measurement '+case)
        rows.append(dict(case=case,round=round_number,ns=float(match[2]),bytes=int(match[3]),allocs=int(match[4])))
    print('round',round_number+1,flush=True)
summary=[]
for case in cases:
    selected=[r for r in rows if r['case']==case]
    summary.append(dict(case=case,median_ns=statistics.median(r['ns'] for r in selected),min_ns=min(r['ns'] for r in selected),max_ns=max(r['ns'] for r in selected),bytes=statistics.median(r['bytes'] for r in selected),allocs=statistics.median(r['allocs'] for r in selected)))
data=dict(library='CMLGo v0.2.3',alternative='Participle v2.1.4 NewSimple + keyword lookup CML adapter',gomaxprocs=1,count=args.rounds,benchtime=args.benchtime,order='rotated per round',summary=summary,raw=rows)
(root/'bench/results/stages.json').write_text(json.dumps(data,indent=2)+'\n')
(root/'bench/results/stages.txt').write_text('\n'.join(raw))
for row in summary:print(row)
