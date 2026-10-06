#!/usr/bin/env python3
"""One JVM serves as the reference oracle; production implementation remains Go."""
import argparse,base64,json,subprocess
from pathlib import Path
from report_paths import portable_report
root=Path(__file__).resolve().parents[1]
p=argparse.ArgumentParser();p.add_argument('--reference-lib',required=True);p.add_argument('--java-bin',required=True);p.add_argument('--directory',default='testdata/conformance');args=p.parse_args()
classes=root/'bin/oracle';classes.mkdir(parents=True,exist_ok=True)
classpath=str(Path(args.reference_lib)/'*')
subprocess.run([str(Path(args.java_bin)/'javac'),'-cp',classpath,'-d',str(classes),str(root/'scripts/oracle/ReferenceOracle.java')],check=True)
files=sorted((root/args.directory).rglob('*.cml'))
run=subprocess.run([str(Path(args.java_bin)/'java'),'-cp',str(classes)+':'+classpath,'ReferenceOracle',*[str(p) for p in files]],capture_output=True,text=True,check=True,timeout=120)
rows=[]
for line in run.stdout.splitlines():
 parts=line.split('\t')
 if len(parts)!=3:continue
 path,status,encoded=parts
 native=subprocess.run([str(root/'bin/cm'),'validate','--json','-i',path],capture_output=True,text=True,timeout=15)
 actual=json.loads(native.stdout)

 oracle=None if status=='crash' else status=='true'
 row=dict(file=str(Path(path).relative_to(root)),reference_valid=oracle,go_valid=actual['valid'],match=oracle==actual['valid'] if oracle is not None else None,go_diagnostics=actual['diagnostics'],reference_errors=base64.b64decode(encoded).decode())
 rows.append(row)
 if row['match'] is False:print('MISMATCH',row['file'],oracle,actual['valid'],flush=True)
report=dict(reference='Context Mapper 6.12.0 Xtext IResourceValidator CheckMode.ALL',files=len(rows),matches=sum(r['match'] is True for r in rows),crashes=sum(r['match'] is None for r in rows),rows=rows)
(root/('bench/results/differential-'+Path(args.directory).name+'.json')).write_text(json.dumps(portable_report(report, root, [Path(args.reference_lib).resolve(), Path(args.java_bin).resolve()]),ensure_ascii=False,indent=2)+'\n')
print({k:v for k,v in report.items() if k!='rows'})
