import { test } from 'node:test'
import assert from 'node:assert/strict'
import { parseShareAddress } from './shares.ts'

test('reads the ways people type a share', () => {
  assert.deepEqual(parseShareAddress('\\\\nas\\Media'), { kind: 'smb', host: 'nas', share: 'Media', folder: '' })
  assert.deepEqual(parseShareAddress('  //nas.local/Media/Films/2019  '), { kind: 'smb', host: 'nas.local', share: 'Media', folder: 'Films/2019' })
  assert.deepEqual(parseShareAddress('smb://192.168.1.20/Media/'), { kind: 'smb', host: '192.168.1.20', share: 'Media', folder: '' })
  assert.deepEqual(parseShareAddress('SMB://[fd00::20]/Media'), { kind: 'smb', host: '[fd00::20]', share: 'Media', folder: '' })
  assert.deepEqual(parseShareAddress('nas:/volume1/media'), { kind: 'nfs', host: 'nas', share: '/volume1/media', folder: '' })
  assert.deepEqual(parseShareAddress('nfs://nas/mnt/user/media'), { kind: 'nfs', host: 'nas', share: '/mnt/user/media', folder: '' })
})

test('refuses what it must', () => {
  const bad = (s: string, has: RegExp) => {
    const r = parseShareAddress(s)
    assert.ok('error' in r, `${s} should be refused`)
    assert.match((r as { error: string }).error, has)
  }
  bad('', /Type the address/)
  bad('smb://me:secret@nas/Media', /their own boxes/)
  bad('\\\\me@nas\\Media', /their own boxes/)
  bad('smb://nas:4455/Media', /port/)
  bad('smb://nas', /Add the share/)
  bad('\\\\nas', /Add the share/)
  bad('nas:/', /exported folder/)
  bad('smb://nas/Media/../etc', /"\.\."/)
  bad('smb://na s/Media', /isn't a computer/)
  bad('smb://nas/Me\u0000dia', /characters/)
  bad('C:\\Media', /Type it as/)
  bad('http://nas/Media', /Type it as/)
  bad('C:/Media', /drive on the server/)
})
