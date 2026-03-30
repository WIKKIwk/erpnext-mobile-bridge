# Werka 通知与最近操作优化计划

## 1. 目标

只优化两块：

- Bildirishnomalar
- So‘nggi harakatlar

别的先不碰。

## 2. 为什么做

现在这两块仍然可能慢。

原因很可能不是写入。

而是读取路径太重。

## 3. 本次只做什么

只做：

- Werka notifications list
- Werka recent/history list
- 如果需要，再看 notification detail

## 4. 不做什么

不做：

- Supplier 最近记录
- Admin 活动页
- 其它无关角色

## 5. 核心原则

和前面一样：

1. 读走 DB
2. 写走 ERP API
3. 第一屏先出来
4. 不能 giant preload

## 6. 现状假设

这两个页面现在慢，很可能因为：

- history 还是 broad collector
- comments / detail 又做了多次读取
- app 打开页面时先等全量数据

## 7. 正确方向

### 通知页

- 先拿第一页
- 分页
- 无限滚动
- 不一次拉完

### 最近页

- 也是先拿第一页
- 后续滚动追加

## 8. 如果 detail 很重

detail 可以后面单独优化。

先把 list 打快。

## 9. 后端任务

- DB-backed recent list
- DB-backed notifications list
- `limit + offset`
- 稳定排序

## 10. 前端任务

- list 打开快
- infinite scroll
- 首屏不等全量

## 11. 验收

- notifications 打开明显快
- recent 打开明显快
- 可以继续下滑加载
- 测试通过

## 12. 一句话总结

这次只做：

`Werka 通知页 + 最近页 = DB直读 + 分页 + 快速首屏`
