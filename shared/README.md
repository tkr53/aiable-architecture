# shared/

What may live here is limited to purely technical, behavior-free elements that belong to no
capability's spec (cross-cutting type definitions, format constants, etc.).

> 日本語版は [README.ja.md](README.ja.md) を参照。

The test is "should this have a spec/test?". If it should, it is behavior, and it must belong to
some capability. The only things you may offload to shared are those that do not need to be tested.

Loosening this policy collapses locality and forces the AI to read all of shared just to fix one
feature. Keep shared deliberately thin. This worked example currently has no shared elements.
