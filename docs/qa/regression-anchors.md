# 回归锚点

> **写入规则**：阶段锁版时，QA Agent 提取本阶段关键验收条件写入此文件。
>
> **执行规则**：后续阶段每次 QA 测试，本文件中的锚点必须全量复验。任一失败 → QA 驳回。

---

## 阶段 1（locked，{date}）

- [ ] 用户注册：POST /api/v1/auth/register → 200 + JWT access token
- [ ] 用户登录：POST /api/v1/auth/login → 200 + JWT
- [ ] 未登录访问需认证接口 → 401
- [ ] 上传视频文件（mp4 ≤500MB）→ 200 + video_id
- [ ] 获取视频详情：GET /api/v1/videos/:id → 200
- [ ] 获取播放地址：GET /api/v1/videos/:id/play-url → 200 + 有效预签名 URL

---

*后续阶段锁版时追加于此*
