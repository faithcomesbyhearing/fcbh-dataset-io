# YAML Generator Selection Logic - Venn Diagrams

This document explains how different argument combinations select different sets of languages using Venn diagrams.

## Overview of Selection Criteria

The YAML generator selects languages based on:
1. **Testament Scope**: N1, N2, O1, O2
2. **Text Format**: USX (`_ET-usx`), Plain (`_ET`)
3. **Stream Format**: HLS (DA filesets), DASH (DA-opus16 filesets)
4. **Content Validation**: `content_loaded = 1`
5. **MMS Support**: Language must be supported by MMS
6. **SA Exclusion**: No corresponding SA fileset exists

## Text Format Selection (USX vs Plain)

```
All Languages with N1DA + N_ET filesets
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │   USX Only          │    │        Plain Only           │ │
│  │   (_ET-usx)         │    │        (_ET)                │ │
│  │                     │    │                             │ │
│  │  ENGNAB, ERVWTC,    │    │    JPNJCB                   │ │
│  │  FRADPI, INDALA,    │    │                             │ │
│  │  TAMWCO, TELGMI     │    │                             │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Both USX and Plain                         │ │
│  │           (Processed in USX run)                        │ │
│  │                                                         │ │
│  │  ABPWBT, ENGNIV, HINBIB, MALBIB,                       │ │
│  │  PORBBS, RUSDPI, SWEBIB, TAMWCO, etc.                  │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Testament Scope Selection (N1 vs N2)

```
All Languages with New Testament DA filesets
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │   N1 Only           │    │        N2 Only              │ │
│  │   (N1DA filesets)   │    │        (N2DA filesets)      │ │
│  │                     │    │                             │ │
│  │  AMHBIB, BNGDIP,    │    │  ACRWB1, BAMLSB,           │ │
│  │  FIN38V, GUJDPI,    │    │  BNGBBS, ENGCEV,           │ │
│  │  HAUDPI, IBOBIB,    │    │  GEOIBT, HATSBH, etc.      │ │
│  │  JAVLAI, KANBIB     │    │                             │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Both N1 and N2                             │ │
│  │           (73 shared languages)                         │ │
│  │                                                         │ │
│  │  ABPWBT, ACCBSG, ADXNVS, AEUWBT,                       │ │
│  │  BNGWTC, BO1CAP, BULBOB, CATBSS, etc.                  │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Stream Format Selection (HLS vs DASH)

```
All Languages with DA filesets
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  ┌─────────────────────┐    ┌─────────────────────────────┐ │
│  │   HLS Only          │    │        DASH Only            │ │
│  │   (DA filesets)     │    │        (DA-opus16)          │ │
│  │                     │    │                             │ │
│  │  Most languages     │    │  Fewer languages            │ │
│  │  have standard      │    │  have opus16                │ │
│  │  HLS filesets       │    │  filesets                   │ │
│  └─────────────────────┘    └─────────────────────────────┘ │
│                                                             │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │              Both HLS and DASH                          │ │
│  │           (Languages with both formats)                 │ │
│  │                                                         │ │
│  │  ENGNIV, PORBBS, RUSDPI, etc.                          │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## Complete Selection Matrix

| Testament | Text Format | Stream | Count | Description |
|-----------|-------------|--------|-------|-------------|
| N1        | USX         | HLS    | 82    | Languages with N1DA + N_ET-usx, no N1SA |
| N1        | Plain       | HLS    | 1     | Languages with N1DA + N_ET, no N1SA, exclude USX |
| N1        | USX         | DASH   | 82    | Same as HLS but with DA-opus16 filesets |
| N1        | Plain       | DASH   | 1     | Same as HLS but with DA-opus16 filesets |
| N2        | USX         | HLS    | 73    | Languages with N2DA + N_ET-usx, no N2SA |
| N2        | Plain       | HLS    | 63    | Languages with N2DA + N_ET, no N2SA, exclude USX |
| O1        | USX         | HLS    | 41    | Languages with O1DA + O_ET-usx, no O1SA |
| O1        | Plain       | HLS    | 1     | Languages with O1DA + O_ET, no O1SA, exclude USX |
| O2        | USX         | HLS    | 19    | Languages with O2DA + O_ET-usx, no O2SA |
| O2        | Plain       | HLS    | 0     | Languages with O2DA + O_ET, no O2SA, exclude USX |

## Key Relationships

### 1. USX vs Plain Text
- **USX run** processes languages that have `_ET-usx` filesets
- **Plain run** processes languages that have `_ET` filesets BUT excludes those that also have `_ET-usx`
- This ensures no language is processed twice

### 2. N1 vs N2
- **N1** finds languages with `N1DA` filesets
- **N2** finds languages with `N2DA` filesets  
- Some languages have both, some have only one

### 3. HLS vs DASH
- **HLS** uses `DA` filesets (e.g., `ABPWBTN1DA`)
- **DASH** uses `DA-opus16` filesets (e.g., `ABPWBTN1DA-opus16`)
- Some languages have both formats available

### 4. Exclusion Logic
- Languages are excluded if they already have corresponding `SA` filesets
- This prevents duplicate HLS/DASH generation
- The tool only processes languages that need SA filesets created

## Example Selection Scenarios

### Scenario 1: N1 USX HLS
```
Languages selected:
- Must have N1DA fileset with content_loaded=1
- Must have N_ET-usx fileset with content_loaded=1  
- Must NOT have N1SA fileset
- Must be supported by MMS
- Result: 82 languages (including ABPWBT, ENGNIV, HINBIB, etc.)
```

### Scenario 2: N1 Plain HLS
```
Languages selected:
- Must have N1DA fileset with content_loaded=1
- Must have N_ET fileset with content_loaded=1
- Must NOT have N_ET-usx fileset (exclusion)
- Must NOT have N1SA fileset
- Must be supported by MMS
- Result: 1 language (only JPNJCB)
```

### Scenario 3: N2 USX HLS
```
Languages selected:
- Must have N2DA fileset with content_loaded=1
- Must have N_ET-usx fileset with content_loaded=1
- Must NOT have N2SA fileset
- Must be supported by MMS
- Result: 73 languages (including ABPWBT, ACCBSG, etc.)
```

## Summary

The YAML generator uses sophisticated filtering logic to ensure:
1. **No duplicates**: Each language is processed only once per text format
2. **Proper scope**: Testament scope matches appropriate fileset patterns
3. **Content validation**: Only processes filesets with actual content
4. **MMS compatibility**: Only processes languages supported by MMS
5. **SA exclusion**: Only processes languages that need SA filesets created

This creates a comprehensive and non-overlapping set of YAML files for HLS/DASH generation across all testament types and text formats.
