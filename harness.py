#!/usr/bin/env python3
"""harness.py — Adversarial evaluation for Curse.

Usage:
    python harness.py <file_or_directory>

Returns exit code 0 (pass) or 1 (fail) with diagnostics on stderr.
The Gateway treats a non-zero exit as feedback for autonomous iteration.
"""

import ast
import os
import re
import sys
import json


SECRET_PATTERNS = [
    r'sk-[A-Za-z0-9]{20,}',
    r'api[-_]?key\s*[:=]\s*["\']?[A-Za-z0-9_\-]{16,}',
    r'password\s*[:=]\s*["\']?[^\s"\']{8,}',
    r'secret\s*[:=]\s*["\']?[^\s"\']{8,}',
    r'token\s*[:=]\s*["\']?[A-Za-z0-9_\-\.]{16,}',
    r'bearer\s+[A-Za-z0-9_\-\.]{16,}',
    r'-----BEGIN (RSA|EC|OPENSSH) PRIVATE KEY-----',
]

BLOCKED_FUNCTIONS = [
    'eval',
    'exec',
    'compile',
    '__import__',
    'os.system',
    'subprocess.call',
    'subprocess.Popen',
    'os.popen',
]

BANNED_IMPORTS = [
    'pickle',
    'cPickle',
    'shelve',
    'marshal',
]

class HarnessResult:
    def __init__(self):
        self.passed = True
        self.errors = []
        self.warnings = []

    def fail(self, msg):
        self.passed = False
        self.errors.append(msg)

    def warn(self, msg):
        self.warnings.append(msg)

    def report(self):
        for w in self.warnings:
            print(f"[WARN] {w}", file=sys.stderr)
        for e in self.errors:
            print(f"[FAIL] {e}", file=sys.stderr)
        if self.passed:
            print(f"[PASS] {len(self.warnings)} warnings, 0 failures", file=sys.stderr)
        else:
            print(f"[FAIL] {len(self.warnings)} warnings, {len(self.errors)} failures", file=sys.stderr)
        return 0 if self.passed else 1


def check_python_file(path, result):
    with open(path, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()

    # Check for secrets
    for pattern in SECRET_PATTERNS:
        matches = re.findall(pattern, content, re.IGNORECASE)
        for m in matches:
            result.fail(f"Secret detected: {m[:20]}... in {path}")

    # Try parsing AST
    try:
        tree = ast.parse(content)
    except SyntaxError as e:
        result.fail(f"Syntax error in {path}: {e}")
        return

    # Check for blocked functions
    for node in ast.walk(tree):
        if isinstance(node, ast.Call):
            if isinstance(node.func, ast.Name):
                if node.func.id in BLOCKED_FUNCTIONS:
                    line = getattr(node, 'lineno', '?')
                    result.fail(f"Blocked function '{node.func.id}' at {path}:{line}")
            elif isinstance(node.func, ast.Attribute):
                full = f"{node.func.value.id}.{node.func.attr}" if isinstance(node.func.value, ast.Name) else ""
                if full in BLOCKED_FUNCTIONS:
                    line = getattr(node, 'lineno', '?')
                    result.fail(f"Blocked call '{full}' at {path}:{line}")

        if isinstance(node, ast.Import):
            for alias in node.names:
                if alias.name in BANNED_IMPORTS:
                    result.fail(f"Banned import '{alias.name}' in {path}")

        if isinstance(node, ast.ImportFrom):
            if node.module in BANNED_IMPORTS:
                result.fail(f"Banned import '{node.module}' in {path}")


def check_go_file(path, result):
    # Basic sanity checks for Go files
    with open(path, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()

    for pattern in SECRET_PATTERNS:
        matches = re.findall(pattern, content, re.IGNORECASE)
        for m in matches:
            result.fail(f"Secret detected: {m[:20]}... in {path}")


def check_generic(path, result):
    with open(path, 'r', encoding='utf-8', errors='replace') as f:
        content = f.read()

    for pattern in SECRET_PATTERNS:
        matches = re.findall(pattern, content, re.IGNORECASE)
        for m in matches:
            result.fail(f"Secret detected: {m[:20]}... in {path}")

    # Check file size
    size = len(content)
    if size > 1024 * 1024:
        result.fail(f"File too large: {size} bytes in {path}")


def main():
    if len(sys.argv) < 2:
        print("Usage: harness.py <file_or_directory>", file=sys.stderr)
        sys.exit(1)

    target = sys.argv[1]
    result = HarnessResult()

    if os.path.isfile(target):
        files = [target]
    elif os.path.isdir(target):
        files = []
        for root, _, filenames in os.walk(target):
            for fn in filenames:
                files.append(os.path.join(root, fn))
    else:
        print(f"Target not found: {target}", file=sys.stderr)
        sys.exit(1)

    for fpath in files:
        if fpath.endswith('.py'):
            check_python_file(fpath, result)
        elif fpath.endswith('.go'):
            check_go_file(fpath, result)
        else:
            check_generic(fpath, result)

    sys.exit(result.report())


if __name__ == '__main__':
    main()
