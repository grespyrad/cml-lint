#!/usr/bin/env python3
"""Воспроизвести Java/Xtext и Go snapshots закреплённого upstream примера."""
import argparse
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--reference-lib', required=True, type=Path)
    parser.add_argument('--java-bin', required=True, type=Path)
    parser.add_argument('--cm', required=True, type=Path)
    parser.add_argument('--output', required=True, type=Path)
    args = parser.parse_args()
    source = ROOT / 'testdata/upstream/insurance-example/Insurance-Example-Stage-3.cml'
    args.output.mkdir(parents=True, exist_ok=True)
    with tempfile.TemporaryDirectory(prefix='cml-formatter-oracle-') as directory:
        classes = Path(directory)
        classpath = str(args.reference_lib.resolve() / '*')
        subprocess.run([str(args.java_bin / 'javac'), '-cp', classpath, '-d', str(classes),
                        str(ROOT / 'scripts/oracle/FormatterOracle.java')], check=True, timeout=60)
        subprocess.run([str(args.java_bin / 'java'), '-cp', str(classes) + ':' + classpath,
                        'FormatterOracle', str(source), str(args.output / 'insurance-stage-3.java.cml')],
                       check=True, timeout=60)
        java_snapshot = args.output / 'insurance-stage-3.java.cml'
        java_snapshot.write_bytes(java_snapshot.read_bytes().rstrip(b'\n') + b'\n')
        pretty = subprocess.run([str(args.cm.resolve()), 'format', str(source)],
                                check=True, capture_output=True, timeout=30).stdout
        (args.output / 'insurance-stage-3.pretty.cml').write_bytes(pretty)
    print('Java/Xtext и Go snapshots сохранены в указанном output-каталоге.')


if __name__ == '__main__':
    main()
