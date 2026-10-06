#!/usr/bin/env python3
"""Собрать исследовательские probes; не участвует в production build."""
import os
import shutil
import subprocess
from pathlib import Path

root = Path(__file__).resolve().parents[1]
env = {**os.environ, 'GOWORK': 'off'}
(root / 'bin').mkdir(exist_ok=True)

def run(command, environment=env):
    subprocess.run(command, cwd=root, env=environment, check=True)

run(['go', 'build', '-o', 'bin/cml-bench', './cmd/cml-bench'])
run(['python3', 'scripts/compile_bench_grammar.py'])
run(['clang++', '-std=c++20', '-O3', '-o', 'bin/cml-syntax-cpp', 'bench/cpp/main.cpp'])
rustup = shutil.which('rustup')
if rustup:
    rustc = subprocess.check_output([rustup, 'which', '--toolchain', 'stable', 'rustc'], text=True).strip()
    run([rustup, 'run', 'stable', 'cargo', 'build', '--manifest-path', 'bench/rust/Cargo.toml', '--release'], {**env, 'RUSTC': rustc})
else:
    run(['cargo', 'build', '--manifest-path', 'bench/rust/Cargo.toml', '--release'])
