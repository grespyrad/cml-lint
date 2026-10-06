#!/usr/bin/env python3
"""Compare standalone Go validation against the official Context Mapper CLI."""
import argparse,json,os,re,subprocess
from pathlib import Path
from report_paths import portable_report
root=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--reference-cli',required=True);p.add_argument('--java-home');p.add_argument('--directory',default='testdata/upstream');args=p.parse_args()
env={**os.environ}
if args.java_home:env['JAVA_HOME']=args.java_home
files=sorted((root/args.directory).rglob('*.cml'))
rows=[]
for file in files:
 original=subprocess.run([args.reference_cli,'validate','-i',str(file)],env=env,capture_output=True,text=True,timeout=30)
 output=original.stdout+'\n'+original.stderr
 oracle=original.returncode==0 and not re.search(r'(^|\n)ERROR|Exception',output)
 native=subprocess.run([str(root/'bin/cm'),'validate','--json','-i',str(file)],capture_output=True,text=True,timeout=30)
 actual=json.loads(native.stdout)

 row=dict(file=str(file.relative_to(root)),reference_valid=oracle,go_valid=actual['valid'],match=bool(oracle)==actual['valid'],go_diagnostics=actual['diagnostics'],reference_output=output,reference_exit=original.returncode)
 rows.append(row)
 if not row['match']:print('MISMATCH',row['file'],row['reference_valid'],row['go_valid'],flush=True)
report=dict(reference='Context Mapper CLI 6.12.0',files=len(rows),matches=sum(r['match'] for r in rows),rows=rows)
(root/('bench/results/differential-'+Path(args.directory).name+'.json')).write_text(json.dumps(portable_report(report, root, [Path(args.reference_cli).resolve().parent, *([Path(args.java_home).resolve()] if args.java_home else [])]),ensure_ascii=False,indent=2)+'\n')
print('Matches',report['matches'],'/',report['files'])
