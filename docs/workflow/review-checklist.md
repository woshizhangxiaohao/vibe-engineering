# Review Checklist (Taste Rules)

## 1. Responsibility Clarity
- [ ] 每个模块是否只有一个清晰职责？
- [ ] 是否存在“为了复用而复用”的抽象？

## 2. Complexity Placement
- [ ] 复杂度是否集中在正确的地方？
- [ ] 有没有把复杂度泄露到调用方？

## 3. Readability Over Cleverness
- [ ] 半年后我还能一眼看懂吗？
- [ ] 有没有“看起来聪明但解释成本很高”的代码？

## 4. Change Surface Area
- [ ] 这次改动影响的范围是否可控？
- [ ] 是否修改了不必要的文件/接口？

## 5. Test Meaningfulness
- [ ] 测试是否验证“行为”，而不是实现？
- [ ] 有没有为了让测试通过而写的测试？

## 6. Future Maintenance
- [ ] 我敢不敢让别人接手这段代码？
- [ ] 出问题时，我能快速定位吗？

## 7. Assumption Explicitness
- [ ] 关键假设是否写在代码或文档中？
- [ ] 是否存在“隐含前提”？

⚠️ 任何一条你犹豫了 = 要继续改
