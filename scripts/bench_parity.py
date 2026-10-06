#!/usr/bin/env python3
"""Verify Go/Rust/C++ recognition parity before comparing wall-clock time."""
import argparse,json,os,subprocess
from pathlib import Path
root=Path(__file__).resolve().parents[1]
engines={
 'go':lambda f:[str(root/'bin/cml-bench'),'-syntax','-input',str(f),'-n','1'],
 'rust':lambda f:[str(root/'bench/rust/target/release/cml-syntax-bench'),str(f),'1'],
 'cpp':lambda f:[str(root/'bin/cml-syntax-cpp'),str(f),'1'],
}
rows=[]
files=[p for folder in ['upstream','conformance','escapes'] for p in sorted((root/'testdata'/folder).rglob('*.cml'))]
for file in files:
 outcomes={lang:subprocess.run(make(file),cwd=root,stdout=subprocess.DEVNULL,stderr=subprocess.DEVNULL,env={**os.environ,'GOMAXPROCS':'1'}).returncode==0 for lang,make in engines.items()}
 match=len(set(outcomes.values()))==1
 rows.append(dict(file=str(file.relative_to(root)),outcomes=outcomes,match=match))
 if not match:print('MISMATCH',file.name,outcomes,flush=True)
report=dict(files=len(rows),matches=sum(r['match'] for r in rows),rows=rows)
(root/'bench/results/syntax-parity.json').write_text(json.dumps(report,indent=2)+'\n')
print(report['matches'],'/',report['files'])
if report['matches']!=report['files']:raise SystemExit(1)
