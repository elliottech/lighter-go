#!/usr/bin/env python3
"""Build and verify signer artifacts together; never copy stale compressed files."""
import argparse
import tempfile
import base64
import gzip
import hashlib
import json
import os
from pathlib import Path
import re
import shutil
import subprocess

os.environ['GOTOOLCHAIN'] = 'go1.26.8'

root = Path(__file__).resolve().parent.parent
wasm = root / 'web-wasm'
parser = argparse.ArgumentParser()
parser.add_argument('--check', action='store_true', help='Rebuild in a temporary directory and compare committed artifacts')
parser.add_argument('--no-copy', action='store_true', help='Do not deliver artifacts to sibling frontend checkouts')
args = parser.parse_args()
temporary = tempfile.TemporaryDirectory() if args.check else None
out = Path(temporary.name) if temporary else wasm

def run(*args, **kwargs):
    return subprocess.check_output(args, cwd=root, **kwargs)

# Every artifact records the exact source tree as well as Git provenance.
sources = sorted([root / 'go.mod', root / 'go.sum', wasm / 'build.py', wasm / 'build.sh', wasm / 'wasm-template.html'] + list((root / 'signer').glob('*.go')) + list(wasm.glob('*.go')) + list((root / 'types').rglob('*.go')))
sources = [path for path in sources if not path.name.endswith('_test.go')]
h = hashlib.sha256()
for path in sources:
    h.update(str(path.relative_to(root)).encode() + b'\0' + path.read_bytes() + b'\0')
run('go', 'build', '-buildvcs=false', '-trimpath', '-ldflags=-s -w', '-o', str(out / 'main.wasm'), './web-wasm', env={**os.environ, 'GOOS': 'js', 'GOARCH': 'wasm'})
binary = (out / 'main.wasm').read_bytes()
(out / 'main.wasm.gz').write_bytes(gzip.compress(binary, mtime=0))
run('brotli', '-Z', '-f', '-k', str(out / 'main.wasm'))
assert gzip.decompress((out / 'main.wasm.gz').read_bytes()) == binary
assert run('brotli', '-d', '-c', str(out / 'main.wasm.br')) == binary
exports = re.findall(r'register\("([^"\n]+)"', (wasm / 'main.go').read_text())
assert len(exports) == len(set(exports))
manifest = {
    'protocol': 2,
    'sourceCommit': run('git', 'rev-parse', 'HEAD').decode().strip(),
    'sourceTreeSha256': h.hexdigest(),
    'goVersion': run('go', 'env', 'GOVERSION').decode().strip(),
    'sha256': hashlib.sha256(binary).hexdigest(),
    'exports': exports,
}
goroot = Path(run('go', 'env', 'GOROOT').decode().strip())
runtime = goroot / 'lib/wasm/wasm_exec.js'
if not runtime.exists(): runtime = goroot / 'misc/wasm/wasm_exec.js'
html = (wasm / 'wasm-template.html').read_text().replace('{{GO_RUNTIME}}', runtime.read_text()).replace('{{WASM_BASE64}}', base64.b64encode(binary).decode()).replace('{{SIGNER_EXPORTS}}', json.dumps(exports))
assert '{{' not in html
script_hashes = ["'sha256-" + base64.b64encode(hashlib.sha256(script.encode()).digest()).decode() + "'" for script in re.findall(r'<script>(.*?)</script>', html, re.S)]
html = html.replace("'unsafe-inline'", ' '.join(script_hashes))
manifest['runtimeSha256'] = hashlib.sha256(runtime.read_bytes()).hexdigest()
manifest['htmlSha256'] = hashlib.sha256(html.encode()).hexdigest()
(out / 'main.manifest.json').write_text(json.dumps(manifest, indent=2) + '\n')
(out / 'sdk.html').write_text(html)
if args.check:
    committed = json.loads((wasm / 'main.manifest.json').read_text())
    # Artifact-only commits change HEAD without changing build inputs. Validate the
    # recorded source commit against those inputs rather than requiring HEAD itself.
    source_commit = committed['sourceCommit']
    assert re.fullmatch(r'[0-9a-f]{40}', source_commit), 'Invalid source commit'
    # A squash merge or shallow checkout may not retain the original build commit.
    # The rebuilt source-tree hash below is mandatory; Git provenance is additionally
    # checked when the recorded object is available, without requiring ancestry.
    available = subprocess.run(['git', 'cat-file', '-e', source_commit + '^{commit}'], cwd=root, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL).returncode == 0
    if available:
        source_hash = hashlib.sha256()
        for path in sources:
            relative = str(path.relative_to(root))
            try:
                recorded = run('git', 'show', source_commit + ':' + relative, stderr=subprocess.PIPE)
            except subprocess.CalledProcessError:
                raise SystemExit('Missing provenance file: ' + relative + ' at source commit ' + source_commit) from None
            source_hash.update(relative.encode() + b'\0' + recorded + b'\0')
        assert source_hash.hexdigest() == committed['sourceTreeSha256'], 'Manifest source provenance mismatch'
    generated = dict(manifest)
    for record in (committed, generated): record.pop('sourceCommit')
    assert committed == generated, 'Committed manifest differs from generated manifest'
    for name in ['main.wasm', 'sdk.html']:
        assert (wasm / name).read_bytes() == (out / name).read_bytes(), 'Stale artifact: ' + name
    assert gzip.decompress((wasm / 'main.wasm.gz').read_bytes()) == binary, 'Stale committed gzip'
    assert run('brotli', '-d', '-c', str(wasm / 'main.wasm.br')) == binary, 'Stale committed Brotli'
    temporary.cleanup()
    print('PASS: committed signer artifacts match source, runtime and provenance')
    raise SystemExit(0)
print(json.dumps(manifest, indent=2))
