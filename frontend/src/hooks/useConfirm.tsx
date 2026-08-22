import { useState, useCallback } from 'react';
import { Modal } from '@arco-design/web-react';

interface ConfirmOptions {
  title?: string;
  content: string;
  okText?: string;
  cancelText?: string;
  okType?: 'primary' | 'secondary' | 'outline' | 'text';
}

export function useConfirm() {
  const [open, setOpen] = useState(false);
  const [options, setOptions] = useState<ConfirmOptions>({
    title: 'Confirm',
    content: '',
    okText: 'Confirm',
    cancelText: 'Cancel',
    okType: 'primary',
  });
  const [resolvePromise, setResolvePromise] = useState<((v: boolean) => void) | null>(null);

  const confirm = useCallback((opts: ConfirmOptions): Promise<boolean> => {
    setOptions(opts);
    setOpen(true);
    return new Promise((resolve) => {
      setResolvePromise(() => resolve);
    });
  }, []);

  const handleOk = useCallback(() => {
    setOpen(false);
    resolvePromise?.(true);
  }, [resolvePromise]);

  const handleCancel = useCallback(() => {
    setOpen(false);
    resolvePromise?.(false);
  }, [resolvePromise]);

  const ConfirmModal = () => (
    <Modal
      visible={open}
      title={options.title}
      onOk={handleOk}
      onCancel={handleCancel}
      okText={options.okText}
      cancelText={options.cancelText}
      autoFocus={false}
    >
      <p>{options.content}</p>
    </Modal>
  );

  return { confirm, ConfirmModal };
}
