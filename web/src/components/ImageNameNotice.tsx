import { useState } from 'react'
import {
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogContentText,
  DialogTitle,
} from '@mui/material'
import { useTranslation } from '../i18n'

// Whether the old image name is still being published is a property of the
// registry, not of the account, so the answer to "stop telling me" belongs in
// this browser and needs no server state. Delete this component, its strings,
// and the flag below together when the old name is retired.
const DISMISSED_STORAGE_KEY = 'tempest-zt.image-name-notice.dismissed'

/** The two storage calls this notice needs, so a test can pass a stub. */
export type NoticeStorage = Pick<Storage, 'getItem' | 'setItem'>

function readDismissed(storage: NoticeStorage | undefined): boolean {
  try {
    return storage?.getItem(DISMISSED_STORAGE_KEY) === 'true'
  } catch {
    // Storage that refuses access (private mode, blocked cookies) still shows
    // the notice; it only cannot remember that it was dismissed.
    return false
  }
}

function rememberDismissed(storage: NoticeStorage | undefined): void {
  try {
    storage?.setItem(DISMISSED_STORAGE_KEY, 'true')
  } catch {
    // Nothing to do: the notice reappears on the next visit.
  }
}

export function imageNameNoticeDismissed(storage: NoticeStorage | undefined = globalThis.localStorage): boolean {
  return readDismissed(storage)
}

interface ImageNameNoticeProps {
  /** Only the administrator deploys the container, so only that account sees it. */
  storage?: NoticeStorage | undefined
}

export default function ImageNameNotice({ storage = globalThis.localStorage }: ImageNameNoticeProps) {
  const { translateText } = useTranslation()
  const [open, setOpen] = useState(() => !readDismissed(storage))

  if (!open) {
    return null
  }

  const closeForNow = () => setOpen(false)
  const closeForGood = () => {
    rememberDismissed(storage)
    setOpen(false)
  }

  return (
    <Dialog
      open
      onClose={closeForNow}
      aria-labelledby="image-name-notice-title"
      maxWidth="sm"
      fullWidth
    >
      <DialogTitle id="image-name-notice-title">
        {translateText('容器镜像名称已更改')}
      </DialogTitle>
      <DialogContent>
        <DialogContentText component="div" sx={{ whiteSpace: 'pre-line' }}>
          {translateText('新名称是 ghcr.io/gt-610/tempest-zt，旧名称 ghcr.io/gt-610/tairitsu 在过渡期内仍会收到更新。')}
          {'\n\n'}
          {translateText('过渡期结束后旧名称不再更新，继续使用也不会报错，只会停留在最后发布的那个版本。')}
          {'\n\n'}
          {translateText('迁移方法：把 compose 文件或 docker run 命令里的 image 改成新名称，然后重新拉取镜像。')}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={closeForNow}>
          {translateText('我知道了')}
        </Button>
        <Button onClick={closeForGood} variant="contained">
          {translateText('不再提示')}
        </Button>
      </DialogActions>
    </Dialog>
  )
}
