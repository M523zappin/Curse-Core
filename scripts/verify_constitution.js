const fs = require('fs');
const path = require('path');

function parseConstitution(filePath) {
  const content = fs.readFileSync(filePath, 'utf-8');
  const lines = content.split('\n');
  const principles = [];
  const rules = [];
  let inPrinciples = false;
  let inRules = false;

  for (const line of lines) {
    const s = line.trim();
    if (s.startsWith('## Principles')) { inPrinciples = true; inRules = false; continue; }
    if (s.startsWith('## Guardrails')) { inPrinciples = false; inRules = true; continue; }
    if (s.startsWith('##')) { inPrinciples = false; inRules = false; }

    if (inPrinciples && /^\d+\./.test(s)) {
      principles.push(s);
    }

    if (inRules && s.startsWith('|') && !s.startsWith('|---') && !s.startsWith('| Rule')) {
      const parts = s.split('|').map(p => p.trim());
      if (parts.length >= 5 && parts[1] && parts[1] !== '-') {
        rules.push({
          id: parts[1],
          check: parts[2],
          severity: parts[3].replace(/`/g, ''),
          description: parts[4]
        });
      }
    }
  }

  console.log('CONSTITUTION.md parsed successfully');
  console.log(`  Principles: ${principles.length}`);
  principles.forEach(p => console.log(`    ${p}`));
  console.log(`  Guardrails:  ${rules.length}`);
  const blocks = rules.filter(r => r.severity === 'block');
  const warns = rules.filter(r => r.severity === 'warn');
  console.log(`    Block rules: ${blocks.length}`);
  console.log(`    Warn rules:  ${warns.length}`);
  rules.forEach(r => console.log(`    [${r.severity}] ${r.id}: ${r.check}`));
  
  if (principles.length === 0) throw new Error('No principles parsed');
  if (rules.length === 0) throw new Error('No guardrails parsed');
  
  console.log('\n✓ Constitution is valid');
}

const filePath = process.argv[2] || 'CONSTITUTION.md';
parseConstitution(path.resolve(filePath));
