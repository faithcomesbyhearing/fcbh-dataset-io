# YAML Generator Selection Logic - Venn Diagrams

This document explains how different argument combinations select different sets of **Bible filesets** using Venn diagrams.

## Overview of Selection Criteria

The YAML generator selects Bible filesets based on:
1. **Testament Scope**: N1, N2, O1, O2
2. **Text Format**: USX (`_ET-usx`), Plain (`_ET`)
3. **Stream Format**: HLS (DA filesets), DASH (DA-opus16 filesets)
4. **Content Validation**: `content_loaded = 1`
5. **MMS Support**: Bible's language must be supported by MMS
6. **SA Exclusion**: No corresponding SA fileset exists

**Important Note**: The generator creates one YAML file per **fileset**, not per language. Since languages can have multiple Bible filesets (e.g., different translations, editions), you may see multiple YAML files for the same language (same first 3 characters of fileset ID).

## Text Format Selection (USX vs Plain)

```
All Bible Filesets with N1DA + N_ET filesets
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│        ┌─────────────────┐        ┌─────────────────┐      │
│        │   Has USX       │        │   Has Plain     │      │
│        │   (_ET-usx)     │        │   (_ET)         │      │
│        │                 │        │                 │      │
│        │  ENGNAB, ERVWTC,│        │    JPNJCB       │      │
│        │  FRADPI, INDALA,│        │                 │      │
│        │  TAMWCO, TELGMI │        │                 │      │
│        └─────────────────┼────────┼─────────────────┘      │
│                         │        │                         │
│                         │        │                         │
│                    ┌────┴────────┴────┐                    │
│                    │   Both USX and   │                    │
│                    │      Plain       │                    │
│                    │ (Processed in    │                    │
│                    │   USX run)       │                    │
│                    │                  │                    │
│                    │ ABPWBT, ENGNIV,  │                    │
│                    │ HINBIB, MALBIB,  │                    │
│                    │ PORBBS, RUSDPI,  │                    │
│                    │ SWEBIB, TAMWCO,  │                    │
│                    │ etc.             │                    │
│                    └──────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

## Testament Scope Selection (N1 vs N2)

```
All Bible Filesets with New Testament DA filesets
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│        ┌─────────────────┐        ┌─────────────────┐      │
│        │   Has N1        │        │   Has N2        │      │
│        │   (N1DA filesets)│       │   (N2DA filesets)│      │
│        │                 │        │                 │      │
│        │  AMHBIB, BNGDIP,│        │  ACRWB1, BAMLSB,│      │
│        │  FIN38V, GUJDPI,│        │  BNGBBS, ENGCEV,│      │
│        │  HAUDPI, IBOBIB,│        │  GEOIBT, HATSBH,│      │
│        │  JAVLAI, KANBIB │        │  etc.           │      │
│        └─────────────────┼────────┼─────────────────┘      │
│                         │        │                         │
│                         │        │                         │
│                    ┌────┴────────┴────┐                    │
│                    │   Both N1 and    │                    │
│                    │       N2         │                    │
│                    │ (hundreds of     │                    │
│                    │ shared filesets) │                    │
│                    │                  │                    │
│                    │ ABPWBT, ACCBSG,  │                    │
│                    │ ADXNVS, AEUWBT,  │                    │
│                    │ BNGWTC, BO1CAP,  │                    │
│                    │ BULBOB, CATBSS,  │                    │
│                    │ etc.             │                    │
│                    └──────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

## Stream Format Selection (HLS vs DASH)

```
All Bible Filesets with DA filesets
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│        ┌─────────────────┐        ┌─────────────────┐      │
│        │   Has HLS       │        │   Has DASH      │      │
│        │   (DA filesets) │        │   (DA-opus16)   │      │
│        │                 │        │                 │      │
│        │  Most filesets  │        │  Fewer filesets │      │
│        │  have standard  │        │  have opus16    │      │
│        │  HLS filesets   │        │  filesets       │      │
│        └─────────────────┼────────┼─────────────────┘      │
│                         │        │                         │
│                         │        │                         │
│                    ┌────┴────────┴────┐                    │
│                    │   Both HLS and   │                    │
│                    │      DASH        │                    │
│                    │ (Filesets with   │                    │
│                    │ both formats)    │                    │
│                    │                  │                    │
│                    │ ENGNIV, PORBBS,  │                    │
│                    │ RUSDPI, etc.     │                    │
│                    └──────────────────┘                    │
└─────────────────────────────────────────────────────────────┘
```

## Complete Selection Matrix

| Testament | Text Format | Stream | Count | Description |
|-----------|-------------|--------|-------|-------------|
| N1        | USX         | HLS    | 711   | Bible filesets with N1DA + N_ET-usx, no N1SA |
| N1        | Plain       | HLS    | 1     | Bible filesets with N1DA + N_ET, no N1SA, exclude USX |
| N1        | USX         | DASH   | 711   | Same as HLS but with DA-opus16 filesets |
| N1        | Plain       | DASH   | 1     | Same as HLS but with DA-opus16 filesets |
| N2        | USX         | HLS    | 855   | Bible filesets with N2DA + N_ET-usx, no N2SA |
| N2        | Plain       | HLS    | 3     | Bible filesets with N2DA + N_ET, no N2SA, exclude USX |
| O1        | USX         | HLS    | 92    | Bible filesets with O1DA + O_ET-usx, no O1SA |
| O1        | Plain       | HLS    | 2     | Bible filesets with O1DA + O_ET, no O1SA, exclude USX |
| O2        | USX         | HLS    | 35    | Bible filesets with O2DA + O_ET-usx, no O2SA |
| O2        | Plain       | HLS    | 0     | Bible filesets with O2DA + O_ET, no O2SA, exclude USX |

## Key Relationships

### 1. USX vs Plain Text
- **USX run** processes Bible filesets that have `_ET-usx` filesets
- **Plain run** processes Bible filesets that have `_ET` filesets BUT excludes those that also have `_ET-usx`
- This ensures no fileset is processed twice

### 2. N1 vs N2
- **N1** finds Bible filesets with `N1DA` filesets
- **N2** finds Bible filesets with `N2DA` filesets  
- Some Bibles have both, some have only one

### 3. HLS vs DASH
- **HLS** uses `DA` filesets (e.g., `ABPWBTN1DA`)
- **DASH** uses `DA-opus16` filesets (e.g., `ABPWBTN1DA-opus16`)
- Some Bible filesets have both formats available

### 4. Exclusion Logic
- Bible filesets are excluded if they already have corresponding `SA` filesets
- This prevents duplicate HLS/DASH generation
- The tool only processes Bible filesets that need SA filesets created

## Example Selection Scenarios

### Scenario 1: N1 USX HLS
```
Bible filesets selected:
- Must have N1DA fileset with content_loaded=1
- Must have N_ET-usx fileset with content_loaded=1  
- Must NOT have N1SA fileset
- Must be supported by MMS (using language tree)
- Result: 711 Bible filesets (including ABPWBT, ENGNIV, HINBIB, etc.)
```

### Scenario 2: N1 Plain HLS
```
Bible filesets selected:
- Must have N1DA fileset with content_loaded=1
- Must have N_ET fileset with content_loaded=1
- Must NOT have N_ET-usx fileset (exclusion)
- Must NOT have N1SA fileset
- Must be supported by MMS (using language tree)
- Result: 1 Bible fileset (only JPNJCB)
```

### Scenario 3: N2 USX HLS
```
Bible filesets selected:
- Must have N2DA fileset with content_loaded=1
- Must have N_ET-usx fileset with content_loaded=1
- Must NOT have N2SA fileset
- Must be supported by MMS (using language tree)
- Result: 855 Bible filesets (including ABPWBT, ACCBSG, etc.)
```

## Summary

The YAML generator uses sophisticated filtering logic to ensure:
1. **No duplicates**: Each Bible fileset is processed only once per text format
2. **Proper scope**: Testament scope matches appropriate fileset patterns
3. **Content validation**: Only processes filesets with actual content
4. **MMS compatibility**: Only processes Bible filesets whose languages are supported by MMS
5. **SA exclusion**: Only processes Bible filesets that need SA filesets created

This creates a comprehensive and non-overlapping set of YAML files for HLS/DASH generation across all testament types and text formats. Since languages can have multiple Bible filesets, you may see multiple YAML files for the same language (identified by the first 3 characters of the fileset ID).
