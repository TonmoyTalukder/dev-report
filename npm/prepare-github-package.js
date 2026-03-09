const fs = require('fs');
const path = require('path');

const sourceDir = __dirname;
const outDir = path.join(sourceDir, '.github-package');

fs.rmSync(outDir, { recursive: true, force: true });
fs.mkdirSync(path.join(outDir, 'bin'), { recursive: true });

for (const relativePath of ['README.md', 'install.js', path.join('bin', 'dev-report.js')]) {
  const from = path.join(sourceDir, relativePath);
  const to = path.join(outDir, relativePath);
  fs.mkdirSync(path.dirname(to), { recursive: true });
  fs.copyFileSync(from, to);
}

const packageJsonPath = path.join(sourceDir, 'package.json');
const packageJson = JSON.parse(fs.readFileSync(packageJsonPath, 'utf8'));
packageJson.name = '@tonmoytalukder/dev-report';
packageJson.scripts = {
  postinstall: packageJson.scripts.postinstall,
};
packageJson.publishConfig = {
  registry: 'https://npm.pkg.github.com'
};

fs.writeFileSync(path.join(outDir, 'package.json'), JSON.stringify(packageJson, null, 2) + '\n');
process.stdout.write(outDir);
