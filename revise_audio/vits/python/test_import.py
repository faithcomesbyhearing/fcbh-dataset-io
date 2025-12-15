#!/usr/bin/env python3
"""
Test script to verify So-VITS-SVC can be imported correctly
"""

import os
import sys

# Add So-VITS-SVC to Python path
SO_VITS_SVC_ROOT = os.environ.get('SO_VITS_SVC_ROOT')
if not SO_VITS_SVC_ROOT:
    print("ERROR: SO_VITS_SVC_ROOT environment variable not set")
    print("Set it to the path where so-vits-svc is cloned, e.g.:")
    print("  export SO_VITS_SVC_ROOT=/Users/jrstear/git/so-vits-svc")
    sys.exit(1)

if not os.path.exists(SO_VITS_SVC_ROOT):
    print(f"ERROR: SO_VITS_SVC_ROOT path does not exist: {SO_VITS_SVC_ROOT}")
    sys.exit(1)

sys.path.insert(0, SO_VITS_SVC_ROOT)

print(f"✓ SO_VITS_SVC_ROOT: {SO_VITS_SVC_ROOT}")
print(f"✓ Path exists: {os.path.exists(SO_VITS_SVC_ROOT)}")

# Try to import So-VITS-SVC
try:
    print("\nAttempting to import So-VITS-SVC...")
    from inference.infer_tool import Svc
    print("✓ Successfully imported Svc class from inference.infer_tool")
    
    # Check if we can see the class methods
    if hasattr(Svc, 'slice_inference'):
        print("✓ Svc.slice_inference method found")
    else:
        print("✗ Svc.slice_inference method NOT found")
    
    if hasattr(Svc, '__init__'):
        print("✓ Svc.__init__ method found")
    else:
        print("✗ Svc.__init__ method NOT found")
    
    print("\n✓ So-VITS-SVC import test PASSED")
    sys.exit(0)
    
except ImportError as e:
    print(f"\n✗ ImportError: {e}")
    print("\nThis might be due to missing dependencies.")
    print("Try installing So-VITS-SVC requirements:")
    print(f"  cd {SO_VITS_SVC_ROOT}")
    print("  pip install -r requirements.txt")
    sys.exit(1)
except Exception as e:
    print(f"\n✗ Unexpected error: {e}")
    import traceback
    traceback.print_exc()
    sys.exit(1)

