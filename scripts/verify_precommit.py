#!/usr/bin/env python3
"""Проверка реального pre-commit: автоисправления и частичное индексирование."""
import argparse
from pathlib import Path
import subprocess
import tempfile

parser = argparse.ArgumentParser(description='Проверка pre-commit на временном Git-репозитории.')
parser.add_argument('--cm', required=True, type=Path)
parser.add_argument('--hook-repo', type=Path)
parser.add_argument('--revision', default='HEAD')
args = parser.parse_args()
binary = args.cm.resolve()
def run(args, cwd, code=0):
    result = subprocess.run(args, cwd=cwd, capture_output=True, text=True, check=False)
    if result.returncode != code:
        raise AssertionError((args, result.returncode, result.stdout, result.stderr))
    return result.stdout

with tempfile.TemporaryDirectory(prefix='cml-precommit-') as tmp:
    root = Path(tmp)
    run(['git','init','-q'],root)
    config = '''repos:
- repo: local
  hooks:
  - id: cml-format
    name: CML format
    entry: "BINARY format -w --"
    language: system
    types: [file]
    files: '\\.cml$'
  - id: cml-format-check
    name: CML format check
    entry: "BINARY format --check --"
    language: system
    types: [file]
    files: '\\.cml$'
'''.replace('BINARY',str(binary))
    if args.hook_repo:
        config = ('repos:\n- repo: "' + str(args.hook_repo.resolve()) + '"\n'
                  '  rev: "' + args.revision + '"\n'
                  '  hooks:\n  - id: cml-format\n    language_version: system\n'
                  '  - id: cml-format-check\n    language_version: system\n')
    (root/'.pre-commit-config.yaml').write_text(config)
    model = root/'first model.cml'
    model.write_text('BoundedContext ServiceIntegration{}\n')
    (root/'second.cml').write_text('BoundedContext ExternalService{}\n')
    run(['git','add','.'],root)
    run(['pre-commit','run','cml-format-check','--all-files'],root,1)
    assert model.read_text() == 'BoundedContext ServiceIntegration{}\n'
    run(['pre-commit','run','--all-files'],root,1)
    assert model.read_text() == 'BoundedContext ServiceIntegration {\n}\n'
    assert run(['git','show',':first model.cml'],root) == 'BoundedContext ServiceIntegration{}\n'
    run(['git','add','.'],root)
    run(['pre-commit','run','--all-files'],root)
    model.write_text('BoundedContext ServiceIntegration {\n}\n\nBoundedContext Unstaged\n')
    run(['pre-commit','run'],root)
    assert 'BoundedContext Unstaged' in model.read_text()
    assert 'BoundedContext Unstaged' not in run(['git','show',':first model.cml'],root)
    (root/'invalid.cml').write_text('BoundedContext Missing{')
    run(['git','add','invalid.cml'],root)
    run(['pre-commit','run'],root,1)
    assert (root/'invalid.cml').read_text() == 'BoundedContext Missing{'
    print('Verified: multiple files, paths with spaces, automatic fixes, clean rerun, partial staging, invalid syntax.')
