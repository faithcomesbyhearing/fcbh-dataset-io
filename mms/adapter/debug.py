import torch
import logging
import psutil


def memoryStatistics(logger, display_label):
    if torch.cuda.is_available():
        allocated = torch.cuda.memory_allocated() / 1e9
        reserved = torch.cuda.memory_reserved() / 1e9
        free = reserved - allocated
        fragmentation = free / reserved if reserved > 0.0 else 0.0
        gpuMax = torch.cuda.max_memory_allocated() / 1e9
        logger.info(f"{display_label}: Allocated: {allocated:.2f}MB, Reserved: {reserved:.2f}GB, "
                 f"Free: {free:.2f}GB, Fragmentation: {fragmentation:.4f}, GPU Max {gpuMax}")
    cpu_mem = psutil.virtual_memory().percent
    logger.info(f"{display_label}: CPU memory {cpu_mem:.1f}%")


def modelMemoryStatistics(logger, model, display_label):
    total_params = sum(p.numel() for p in model.parameters())
    trainable_params = sum(p.numel() for p in model.parameters() if p.requires_grad)
    logger.info(f"{display_label}:  Total parameters: {total_params / 1e9:.2f}GB")
    logger.info(f"Trainable parameters: {trainable_params / 1e6:.2f}MB")
    logger.info(f"Frozen parameters: {(total_params - trainable_params) / 1e9:.2f}GB")