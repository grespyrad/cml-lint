#!/usr/bin/env python3
"""Повторные syntax benchmarks; сырые замеры сохраняются локально."""
import argparse, json, os, platform, statistics, subprocess, time
from pathlib import Path
root=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--rounds',type=int,default=7);p.add_argument('--iterations',type=int,default=200);args=p.parse_args()
engines={
 'go': [str(root/'bin/cml-bench'),'-syntax','-input'],
 'rust': [str(root/'bench/rust/target/release/cml-syntax-bench')],
 'cpp': [str(root/'bin/cml-syntax-cpp')],
}
env={**os.environ,'GOMAXPROCS':'1'}
results=[]
for file in ['testdata/reference/service-before.cml','testdata/reference/service.cml']:
 for round in range(args.rounds):
  order=list(engines);order=order[round%3:]+order[:round%3]
  for language in order:
   command=engines[language]+[str(root/file)]+(['-n',str(args.iterations)] if language=='go' else [str(args.iterations)])
   begin=time.perf_counter_ns()
   run=subprocess.run(command,cwd=root,env=env,capture_output=True,text=True,check=True)
   sample=json.loads(run.stdout);sample.update(file=file,round=round,process_ns=time.perf_counter_ns()-begin);results.append(sample)
# Process startup+warmup parse+one timed parse, separately from in-memory loop.
startup=[]
for round in range(args.rounds):
 for language in engines:
  cmd=engines[language]+[str(root/'testdata/reference/service.cml')]+(['-n','1'] if language=='go' else ['1'])
  begin=time.perf_counter_ns();subprocess.run(cmd,cwd=root,env=env,stdout=subprocess.DEVNULL,check=True)
  startup.append(dict(language=language,round=round,ns=time.perf_counter_ns()-begin))
summary=[]
for file in ['testdata/reference/service-before.cml','testdata/reference/service.cml']:
 for language in engines:
  samples=[r['ns_per_op'] for r in results if r['file']==file and r['language']==language]
  summary.append(dict(file=file,language=language,median_ns=statistics.median(samples),min_ns=min(samples),max_ns=max(samples)))
data=dict(cmlgo_version=subprocess.check_output(['go','list','-m','-f','{{.Version}}','github.com/grespyrad/CMLGo'],cwd=root,env={**env,'GOWORK':'off'},text=True).strip(),grammar_mode='Original CML 6.12.0',date='2026-10-05',platform=platform.platform(),machine=platform.machine(),gomaxprocs=1,rounds=args.rounds,iterations=args.iterations,summary=summary,raw=results,startup=startup)
(root/'bench/results/comparison.json').write_text(json.dumps(data,ensure_ascii=False,indent=2)+'\n')
for s in summary:print(s)
for lang in engines:print('startup',lang,statistics.median([s['ns'] for s in startup if s['language']==lang])/1e6,'ms')
