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
  const { t, translateText } = useTranslation()
  const memberName = selectedMember?.name || selectedMember?.id || translateText('未知')

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>{translateText('确认移除成员')}</DialogTitle>
      <DialogContent>
        <DialogContentText>
          {t('network.removeMemberConfirm', { name: memberName })}
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
