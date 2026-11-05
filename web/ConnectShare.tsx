import { useEffect, useRef, useState } from 'react'
import { parseShareAddress, type Share, type Shares } from './shares'

// ConnectShare asks for a share's address and, for SMB, a user name and
// password, and has the host connect to it. The password is only ever in this
// form's state, and is emptied after each try.
export default function ConnectShare({ shares, onConnected, onCancel }: { shares: Shares; onConnected: (s: Share) => void; onCancel: () => void }) {
  const [address, setAddress] = useState('')
  const [user, setUser] = useState('')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const first = useRef<HTMLInputElement | null>(null)
  useEffect(() => { first.current?.focus() }, [])

  const parsed = address.trim() ? parseShareAddress(address) : null
  const kind = parsed && !('error' in parsed) ? parsed.kind : null

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    const p = parseShareAddress(address)
    if ('error' in p) { setErr(p.error); first.current?.focus(); return }
    setBusy(true); setErr('')
    try {
      const s = await shares.connect(p.kind === 'smb' ? { ...p, user: user.trim(), password } : p)
      onConnected(s)
    } catch (e) {
      setErr((e as Error).message || "The server couldn't connect to that share.")
    } finally {
      setPassword(''); setBusy(false)
    }
  }

  return (
    <form className="fx-connect" onSubmit={submit} aria-labelledby="fx-connect-title" autoComplete="off"
      onKeyDown={(e) => { if (e.key === 'Escape' && !busy) { e.stopPropagation(); onCancel() } }}>
      <h2 id="fx-connect-title">Connect to a share</h2>
      <p className="muted">The server connects to it itself, so nothing has to be set up on the computer or in Docker. It only reads from the share.</p>
      <label className="fx-field">
        <span>Address</span>
        <input ref={first} className="input" value={address} onChange={(e) => setAddress(e.target.value)} spellCheck={false}
          placeholder="\\nas\Media, smb://nas/Media or nas:/volume1/media" aria-describedby="fx-connect-kind" readOnly={busy} />
      </label>
      <p id="fx-connect-kind" className="muted fx-connect-kind">
        {kind === 'smb' ? 'SMB: the kind of share Windows and most NAS boxes offer.' : kind === 'nfs' ? 'NFS: the kind of share Linux and many NAS boxes offer.' : 'SMB or NFS, read from how the address is written.'}
      </p>
      {kind === 'smb' && (
        <div className="fx-connect-pair">
          <label className="fx-field">
            <span>User name</span>
            <input className="input" value={user} onChange={(e) => setUser(e.target.value)} spellCheck={false} autoCapitalize="off" readOnly={busy} />
          </label>
          <label className="fx-field">
            <span>Password</span>
            <input className="input" type="password" value={password} onChange={(e) => setPassword(e.target.value)} autoComplete="new-password" readOnly={busy} />
          </label>
        </div>
      )}
      {kind === 'smb' && <p className="muted">Use an account on the NAS that may only read the media folders. The server keeps the password encrypted and never shows it again.</p>}
      {kind === 'nfs' && <p className="muted">NFS has no password: any computer on the network can claim to be anyone. What keeps the share safe is the NAS's own setting for which computers may connect, so allow only this server, and read only.</p>}
      {err && <div className="error" role="alert">{err}</div>}
      <div className="fx-connect-buttons">
        <button type="button" className="btn ghost" onClick={onCancel} disabled={busy}>Cancel</button>
        <button type="submit" className="btn primary" disabled={busy || !address.trim()}>{busy ? 'Connecting…' : 'Connect'}</button>
      </div>
    </form>
  )
}
