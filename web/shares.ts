// Network shares (SMB and NFS) that the host connects to itself, in user space,
// so nothing has to be mounted by the operating system. The window only asks:
// it never keeps a password, and the module's helper never reaches the network.

// Share is a connection the host keeps; path is what list() opens.
export type Share = { id: string; name: string; kind: 'smb' | 'nfs'; path: string }

// ShareRequest is what connect() is given. For NFS, share is the whole
// exported path and there is no user or password.
export type ShareRequest = { kind: 'smb' | 'nfs'; host: string; share: string; folder: string; user?: string; password?: string }

// Shares is what a host that can connect to shares passes to the window.
export type Shares = {
  list: () => Promise<Share[]>
  connect: (r: ShareRequest) => Promise<Share>
  forget: (id: string) => Promise<void>
}

export type ShareAddress = Omit<ShareRequest, 'user' | 'password'>

const HOST = /^(?:[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?(?:\.[A-Za-z0-9](?:[A-Za-z0-9-]{0,61}[A-Za-z0-9])?)*|\[[0-9A-Fa-f:.]+\])$/
const BAD_PART = /[\u0000-\u001f\u007f]/

// parseShareAddress reads what people type for a share: \\nas\Media,
// //nas/Media, smb://nas/Media/Films, nas:/volume1/media or
// nfs://nas/volume1/media. It returns the parts, or a sentence saying what's
// wrong. A user name or password in the address is refused, so one never ends
// up in a path, the recently-chosen list or the browser's storage.
export function parseShareAddress(input: string): ShareAddress | { error: string } {
  const s = input.trim()
  if (!s) return { error: 'Type the address of the share.' }
  if (BAD_PART.test(s)) return { error: "That address has characters a share's can't." }
  let kind: 'smb' | 'nfs', host: string, parts: string[]
  const url = /^(smb|nfs):\/\/([^/]*)(\/.*)?$/i.exec(s)
  const unc = /^(?:\\\\|\/\/)([^\\/]+)[\\/]?(.*)$/.exec(s)
  const nfs = /^([^:/\\@]+):(\/(?!\/).*)$/.exec(s)
  if (url) {
    kind = url[1].toLowerCase() as 'smb' | 'nfs'
    host = url[2]
    parts = (url[3] || '').split('/')
  } else if (unc) {
    kind = 'smb'
    host = unc[1]
    parts = unc[2].split(/[\\/]/)
  } else if (nfs && /^[A-Za-z]$/.test(nfs[1])) {
    return { error: 'That is a drive on the server, not a share. Open it from the places instead.' }
  } else if (nfs) {
    kind = 'nfs'
    host = nfs[1]
    parts = nfs[2].split('/')
  } else {
    return { error: 'Type it as \\\\nas\\Media, smb://nas/Media or nas:/volume1/media.' }
  }
  if (host.includes('@')) return { error: 'Put the user name and password in their own boxes, not in the address.' }
  if (/:\d+$/.test(host) && !host.startsWith('[')) return { error: 'Leave out the port: the server uses the usual one for each kind of share.' }
  if (!HOST.test(host)) return { error: `"${host}" isn't a computer's name or address.` }
  parts = parts.filter(Boolean)
  if (parts.some((p) => p === '.' || p === '..')) return { error: 'An address can\'t have "." or ".." in it.' }
  if (!parts.length) return { error: kind === 'smb' ? 'Add the share after the computer, as in \\\\nas\\Media.' : 'Add the exported folder, as in nas:/volume1/media.' }
  if (kind === 'nfs') return { kind, host, share: '/' + parts.join('/'), folder: '' }
  return { kind, host, share: parts[0], folder: parts.slice(1).join('/') }
}
