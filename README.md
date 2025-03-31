# 食堂AB餐管理系统

## 项目简介
食堂AB餐管理系统是一个在上海学校AB餐政策出台后为更好管理AB餐的选择和领取而开发的系统。支持学生选餐、食堂管理、钉钉集成等功能。该系统针对AB餐管理的实际需求设计，提供了完整的餐食管理、用户管理以及选餐订单管理功能。

请注意，本项目仅可用于校园教育免费使用，不可商用。


## 主要功能
- **用户管理**：支持管理员、食堂工作人员、学生等多种角色
- **餐食管理**：创建、更新、删除餐食，设置选餐时间和生效时间
- **选餐系统**：学生可以选择A餐或B餐，支持批量选餐
- **钉钉集成**：支持接入钉钉工作台与钉钉登录
- **消息通知**：自动发送选餐提醒和选餐结果通知
- **取餐扫码**：食堂工作人员通过扫码确认学生取餐
- **统计功能**：统计全校AB餐/班级AB餐人数（在学生管理中筛选对应班级并全选可见）
- **数据导出**：支持将数据导出为Excel表格

## 预览截图
![image](https://github.com/user-attachments/assets/83cada5b-2f5b-405d-9bc4-dd127dd2ab09)
![image](https://github.com/user-attachments/assets/00b1b031-69d4-486d-af6b-1fec8124ec45)
![image](https://github.com/user-attachments/assets/0ba32d70-a374-408b-9aa6-556a36770475)
![image](https://github.com/user-attachments/assets/4dbd398b-c6cf-4adf-8aae-d97e8d2d977a)
![image](https://github.com/user-attachments/assets/d42dc2dc-210e-48d2-a11e-fd36ddf070ec)
![image](https://github.com/user-attachments/assets/32e2ede4-28b1-41cd-ac3c-4149bb71ec44)
![image](https://github.com/user-attachments/assets/da440e7a-c041-445f-b784-47e049a4dda7)
![image](https://github.com/user-attachments/assets/9c380655-2a09-4105-834e-e3fa4a04d75d)




## 项目结构
- 后端：https://github.com/itsHenry35/canteen-management-system
- 前端（管理+学生选餐）：https://github.com/itsHenry35/canteen-management-system-frontend
- 食堂阿姨安卓端：https://github.com/itsHenry35/CanteenClient

## 安装与运行

### 前置要求
- Go 1.16或更高版本
- Node.js 14.0或更高版本
- Android Studio 4.0或更高版本（用于编译安卓客户端）
- Nginx（用于反向代理配置）
- 钉钉组织管理员账号（用于钉钉集成）

### 安装步骤

#### 1. 编译前端
```bash
# 克隆前端仓库
git clone https://github.com/itsHenry35/canteen-management-system-frontend.git

# 进入项目目录
cd canteen-management-system-frontend

# 安装依赖
npm install

# 构建项目
npm run build
```

编译后的文件将位于`build`文件夹中。**重要：请将`static`文件夹中的`js`和`css`文件夹复制到`build`文件夹中。**

#### 2. 编译后端
```bash
# 克隆后端仓库
git clone https://github.com/itsHenry35/canteen-management-system.git

# 进入项目目录
cd canteen-management-system

# 修改钉钉回调配置
# 打开 api-handlers-admin.go 和 api-handlers-meal.go 文件
# 将文件中的 SingleURL 修改为 "你的域名/dingtalk_auth"

# 编译后端
go build
```

#### 3. 部署系统
```bash
# 在服务器上创建必要的目录结构
mkdir -p /path/to/deployment/database/migrations

# 复制编译后的后端可执行文件到部署目录
cp canteen-management-system /path/to/deployment/

# 复制前端构建文件到部署目录
cp -r canteen-management-system-frontend/build/* /path/to/deployment/frontend/

# 复制数据库迁移文件
cp canteen-management-system/database-migrations-schema.sql /path/to/deployment/database/migrations/

# 运行系统（初始管理员账号和密码将在终端中显示）
cd /path/to/deployment
./canteen-management-system
```

#### 4. 配置系统
1. 配置Nginx反向代理，将域名映射到系统默认的8080端口
2. 使用初始管理员账号登录系统
3. 在用户管理中添加食堂工作人员账号
4. 在学生管理中批量导入学生数据
   - 建议基于钉钉生成的家校通讯录表格进行修改
   - 使用Excel预先准备好学生数据
5. 批量生成学生二维码并打印
6. 按照页面指引配置菜单管理

#### 5. 配置钉钉集成
1. 使用组织管理员账号登录[钉钉开放平台](https://open-dev.dingtalk.com/)
2. 选择"应用开发"-"创建应用"
![image](https://github.com/user-attachments/assets/dd56bea2-72aa-490e-b161-8cd95a2e5ec3)
3. 在权限管理中授予以下权限：
   - 【敏感】钉钉教育家校通讯录读权限
   - 通讯录个人信息读权限
4. 在安全设置中配置：
   - 重定向URL（回调域名）：`你的域名/dingtalk_auth`
   - 端内免登地址：`你的域名/dingtalk_auth`
5. 在"添加应用能力"中添加"网页应用"能力，并点击配置：
   - 应用首页地址：`你的域名/dingtalk_auth`
   - PC端首页地址：`你的域名/dingtalk_auth`
6. 在"凭证与基础信息"中：
   - 开启鸿蒙系统适配选项
   - 复制AgentId、AppKey、AppSecret
7. 点击右上角头像，复制CorpId
![image](https://github.com/user-attachments/assets/32b50046-146e-4107-a6c5-2165e30b5d0f)
8. 在版本管理与发布中创建新版本
9. 在食堂选餐管理系统设置中：
   - 填写上述复制的四个值
   - 先点击"保存设置"，再点击"重建映射管理"

#### 6. 配置安卓扫码系统
```bash
# 克隆安卓客户端仓库
git clone https://github.com/itsHenry35/CanteenClient

# 修改API基础URL
# 打开 app/src/main/java/com/itshenry/canteenclient/api/RetrofitClient.kt
# 将 BASE_URL 修改为你的域名

# 使用Android Studio打开项目进行编译
# 将编译好的APK文件分发给食堂工作人员使用的设备
```

## 使用说明
成功部署后，管理员可以登录系统进行以下操作：
1. 管理用户和学生信息
2. 设置每日菜单
3. 查看选餐统计和取餐记录

学生可以通过钉钉进行选餐（学生仅能看见正在进行的选餐与他们参与过的选餐），食堂工作人员可以使用安卓端APP扫描学生二维码确认取餐。

## 技术支持
如遇到技术问题，请通过以下方式联系开发者：
- QQ: 2671230065 
- 微信: itshenryz 
- 邮箱：zhr0305@outlook.com

## 许可说明
本项目仅可用于校园教育免费使用，严禁商业用途。
