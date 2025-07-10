from transformers import TrainerCallback
import torch
import psutil

class MemoryCallback(TrainerCallback):
    def on_step_end(self, args, state, control, **kwargs):
        if state.global_step % 1 == 0:  # Log every 1 steps
            if torch.cuda.is_available():
                allocated = torch.cuda.memory_allocated() / 1024**3
                reserved = torch.cuda.memory_reserved() / 1024**3
                free = reserved - allocated
                fragmentation = free / reserved if reserved > 0 else 0
                gpuMax = torch.cuda.max_memory_allocated() / 1024**3
                print(f"Step: {state.global_step}  Allocated: {allocated:.2f}MB, Reserved: {reserved:.2f}GB, "
                         f"Free: {free:.2f}GB, Fragmentation: {fragmentation:.4f}, GPU Max {gpuMax}")
                if fragmentation > 0.3:  # 30% fragmentation
                    torch.cuda.empty_cache()
                    print("Cleared CUDA cache due to fragmentation", file=sys.stderr)
            cpu_mem = psutil.virtual_memory().percent
            print(f"Step {state.global_step}: CPU memory {cpu_mem:.1f}%")