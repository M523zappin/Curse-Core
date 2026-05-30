#!/usr/bin/env python3
"""Verify that CONSTITUTION.md is parseable and all guardrails are recognized."""

import re
import sys


def parse_constitution(path):
    with open(path, 'r') as f:
        content = f.read()

    lines = content.split('\n')
    in_principles = False
    in_rules = False
    principles = []
    rules = []

    for line in lines:
        stripped = line.strip()
        if stripped.startswith('## Principles'):
            in_principles = True
            in_rules = False
            continue
        if stripped.startswith('## Guardrails'):
            in_principles = False
            in_rules = True
            continue
        if stripped.startswith('##'):
            in_principles = False
            in_rules = False

        if in_principles and re.match(r'^\d+\.', stripped):
            principles.append(stripped)

        if in_rules and stripped.startswith('|') and not stripped.startswith('|---') and not stripped.startswith('| Rule'):
            parts = [p.strip() for p in stripped.split('|')]
            if len(parts) >= 5 and parts[1] and parts[1] != '-':
                rules.append({
                    'id': parts[1],
                    'check': parts[2],
                    'severity': parts[3],
                    'description': parts[4],
                })

    print(f"CONSTITUTION.md parsed successfully")
    print(f"  Principles: {len(principles)}")
    for p in principles:
        print(f"    {p}")

    print(f"  Guardrails:  {len(rules)}")
    blocks = [r for r in rules if r['severity'] == 'block' or r['severity'] == '`block`']
    warns = [r for r in rules if r['severity'] == 'warn' or r['severity'] == '`warn`']
    print(f"    Block rules: {len(blocks)}")
    print(f"    Warn rules:  {len(warns)}")
    for r in rules:
        print(f"    [{r['severity'].replace('`','')}] {r['id']}: {r['check']}")

    if len(principles) == 0:
        print("ERROR: No principles parsed!", file=sys.stderr)
        return False
    if len(rules) == 0:
        print("ERROR: No guardrails parsed!", file=sys.stderr)
        return False

    return True


if __name__ == '__main__':
    ok = parse_constitution(sys.argv[1] if len(sys.argv) > 1 else 'CONSTITUTION.md')
    sys.exit(0 if ok else 1)
