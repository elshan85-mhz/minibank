# PLAN.md — MiniBank

## My decisions

### 1. WHAT

یک سرویس دفتر کل (ledger) REST با Go برای یک استارتاپ فین‌تک که حساب‌ها را نگه می‌دارد، واریز، برداشت و انتقال انجام می‌دهد و تاریخچه تراکنش‌ها را برمی‌گرداند.
هدف اصلی این است که حتی با صدها درخواست همزمان، یک ریال هم گم یا ساخته نشود.

### 2. STRUCTURE

- cmd/minibank/main.go: فقط سیم‌کشی (خواندن config، انتخاب store، ساخت سرور، خاموشی مرتب).
- internal/bank: دامنه (Account و Transaction، اعتبارسنجی، خطاهای ثابت).
- internal/store: اینترفیس Store با دو پیاده‌سازی memory و postgres.
- internal/api: handlerها، middleware و نگاشت خطا به status code.
- scripts/, k8s/, .github/workflows/: بخش DevOps.

### 3. DATA

جدول accounts: id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY، owner TEXT NOT NULL، balance BIGINT NOT NULL DEFAULT 0 CHECK (balance >= 0)، created_at TIMESTAMPTZ NOT NULL DEFAULT now().
جدول transactions: id BIGINT identity، type TEXT (deposit/withdraw/transfer)، id_from و id_to از نوع BIGINT NULL با کلید خارجی به accounts، amount BIGINT CHECK (amount > 0)، created_at.
همه چیز BIGINT است تا شناسه‌های بزرگ مثل 999999999999 خطای 500 نسازند.

### 4. CONCURRENCY

در store درون‌حافظه‌ای یک sync.Mutex دارم و Transfer فقط یک بار قفل می‌گیرد و هر دو حساب را زیر همان قفل تغییر می‌دهد (بدون صدا زدن Withdraw و Deposit، تا بن‌بست یا غیراتمیک شدن پیش نیاید).
در PostgreSQL برداشت با یک UPDATE شرطی (WHERE balance >= $1) انجام می‌شود. Transfer داخل یک تراکنش SQL است: اول با SELECT ... FOR UPDATE هر دو ردیف قفل می‌شوند (404 قبل از 409)، بعد موجودی بررسی می‌شود، سپس دو UPDATE و ثبت ردیف تراکنش.

### 5. SECURITY

همه کوئری‌ها پارامتری ($1، $2) هستند و هرگز با fmt.Sprintf ساخته نمی‌شوند (جلوگیری از SQL injection).
رمزها فقط از env یا Secret می‌آیند و هیچ رمزی در کد Go یا compose.yaml نیست؛ فایل .env در gitignore است.
کانتینر با کاربر غیر root (65532) روی ایمیج scratch اجرا می‌شود، و حجم بدنه درخواست با MaxBytesReader به ۱ مگابایت محدود است.

### 6. OUT OF SCOPE

احراز هویت و مجوز، چند ارزی، صفحه‌بندی لیست‌ها، ابزار جدا برای migration دیتابیس (schema موقع شروع برنامه ساخته می‌شود)، و TLS.

### 7. RISKS

Kubernetes (probeها، Secret، لود ایمیج در kind) احتمالاً زمان بیشتری می‌گیرد. اگر گیر کردم، اول Compose را کامل و سبز نگه می‌دارم تا برنامه حداقل با آن قابل تصحیح باشد و مشکل را در README یادداشت می‌کنم.
ریسک دوم منطق Transfer در PostgreSQL است؛ اگر زمان کم آمد، اول نسخه ساده با قفل ردیف را کامل می‌کنم.

### 8. AI PLAN

از Claude برای توضیح مفاهیم، اسکلت کد و بررسی خطاها استفاده می‌کنم.
هر کدی را که می‌دهد خودم اجرا می‌کنم و با تست و curl امتحان می‌کنم. مخصوصاً بخش‌های پرخطر (Transfer، شمارش rune، ترتیب خطاها، probeها) را خط‌به‌خط می‌خوانم تا بتوانم در دفاع توضیح بدهم.
