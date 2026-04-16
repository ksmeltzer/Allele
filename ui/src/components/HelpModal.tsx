import React, { useState } from 'react';
import IconButton from '@mui/material/IconButton';
import HelpOutlineIcon from '@mui/icons-material/HelpOutlineOutlined';
import CloseIcon from '@mui/icons-material/Close';
import Dialog from '@mui/material/Dialog';
import DialogContent from '@mui/material/DialogContent';
import DialogTitle from '@mui/material/DialogTitle';

export interface HelpModalProps {
  title: string;
  children: React.ReactNode;
  iconColor?: string;
  size?: "small" | "medium" | "large";
}

export default function HelpModal({ title, children, iconColor = "#A6B0C3", size = "small" }: HelpModalProps) {
  const [open, setOpen] = useState(false);

  const handleClickOpen = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen(true);
  };

  const handleClose = (e: React.MouseEvent) => {
    e.stopPropagation();
    setOpen(false);
  };

  return (
    <>
      <IconButton 
        onClick={handleClickOpen}
        size={size}
        sx={{ color: iconColor, '&:hover': { color: 'white', backgroundColor: 'rgba(255,255,255,0.1)' } }}
        aria-label="help"
      >
        <HelpOutlineIcon fontSize={size === "small" ? "inherit" : "small"} />
      </IconButton>

      <Dialog
        open={open}
        onClose={handleClose}
        aria-describedby="help-dialog-description"
        sx={{
          '& .MuiDialog-paper': {
            backgroundColor: '#1B2028',
            color: '#E2E8F0',
            border: '1px solid #2B3139',
            boxShadow: '0 25px 50px -12px rgba(0, 0, 0, 0.5)',
            minWidth: '400px'
          }
        }}
      >
        <DialogTitle sx={{ m: 0, p: 2, pb: 1, borderBottom: '1px solid #2B3139', display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <span className="font-bold tracking-wide">{title}</span>
          <IconButton
            aria-label="close"
            onClick={handleClose}
            sx={{ color: '#A6B0C3', '&:hover': { color: 'white' } }}
          >
            <CloseIcon fontSize="small" />
          </IconButton>
        </DialogTitle>
        <DialogContent sx={{ p: 3, pt: 3 }}>
          <div id="help-dialog-description" className="text-sm text-[#A6B0C3] space-y-4 leading-relaxed">
            {children}
          </div>
        </DialogContent>
      </Dialog>
    </>
  );
}
