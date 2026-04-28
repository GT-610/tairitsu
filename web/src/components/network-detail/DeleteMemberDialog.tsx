import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogTitle } from '@mui/material'
import { useTranslation } from '../../i18n'
import type { NetworkMemberDevice } from './types'

interface DeleteMemberDialogProps {
  open: boolean;
  saving: boolean;
  selectedMember: NetworkMemberDevice | null;
  onClose: () => void;
  onConfirm: () => void;
}

function DeleteMemberDialog({ open, saving, selectedMember, onClose, onConfirm }: DeleteMemberDialogProps) {
  const { translateText } = useTranslation()

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>{translateText('确认移除成员')}</DialogTitle>
      <DialogContent>
        <DialogContentText>
          {translateText('确定要将成员')} "{selectedMember?.name || selectedMember?.id}" {translateText('从网络中移除吗？此操作会删除该成员在当前网络中的记录。')}
        </DialogContentText>
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>{translateText('取消')}</Button>
        <Button onClick={onConfirm} color="error" disabled={saving}>{translateText('移除')}</Button>
      </DialogActions>
    </Dialog>
  )
}

export default DeleteMemberDialog
