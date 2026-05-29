<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { useAdminStore } from '@/stores/admin'
import type { PlatformConfigRecord } from '@/data/adminMock'

const store = useAdminStore()
const isSaving = ref(false)
const form = ref<PlatformConfigRecord>({
  id: 0,
  lineContactUrl: '',
  faqHtml: '',
  shippingFee: 90,
  freeShippingThreshold: 1200,
  featuredCategoryIds: [],
  featuredBrandIds: [],
})

onMounted(async () => {
  if (!store.platformConfig.id) {
    await store.fetchPlatformConfig()
  }
  form.value = { ...store.platformConfig }
})

async function savePlatformConfig() {
  isSaving.value = true
  try {
    const updated = await store.updatePlatformConfig({
      id: form.value.id,
      lineContactUrl: form.value.lineContactUrl.trim(),
      faqHtml: form.value.faqHtml,
      shippingFee: Number.isFinite(form.value.shippingFee) ? form.value.shippingFee : 90,
      freeShippingThreshold: Number.isFinite(form.value.freeShippingThreshold) ? form.value.freeShippingThreshold : 0,
      featuredCategoryIds: form.value.featuredCategoryIds,
      featuredBrandIds: form.value.featuredBrandIds,
    })
    form.value = { ...updated }
    ElMessage.success('平台配置已成功儲存')
  } catch (error) {
    console.error('儲存平台配置失敗:', error)
    ElMessage.error('儲存平台配置失敗: ' + (error as Error).message)
  } finally {
    isSaving.value = false
  }
}
</script>

<template>
  <section class="platform-settings">
    <el-card>
      <template #header>
        <div class="card-header">
          <div>
            <span class="title">平台配置</span>
            <p class="subcopy">統一設定前台全站共用的客服超連結、常見問題與運費規則。</p>
          </div>
        </div>
      </template>

      <el-form label-width="120px">
        <el-form-item label="LINE 超連結">
          <el-input
            v-model="form.lineContactUrl"
            type="url"
            placeholder="例如：https://line.me/R/ti/p/@your-line-id"
          />
          <small>前台右下角客服按鈕將直接跳轉到這個網址。</small>
        </el-form-item>

        <el-form-item label="常見問題 HTML">
          <el-input
            v-model="form.faqHtml"
            type="textarea"
            :rows="10"
            placeholder="可輸入 HTML，例如：&lt;p&gt;Q：多久出貨？&lt;/p&gt;&lt;p&gt;A：確認訂單後 1-2 個工作天內出貨。&lt;/p&gt;"
          />
          <small>商品詳情頁會顯示這段常見問題內容，支援 HTML。</small>
        </el-form-item>

        <div class="shipping-grid">
          <el-form-item label="基礎運費">
            <el-input-number
              v-model="form.shippingFee"
              :min="0"
              :step="1"
              style="width: 100%"
            />
            <small>未達免運門檻時，前台會收取這筆運費。</small>
          </el-form-item>

          <el-form-item label="免運門檻">
            <el-input-number
              v-model="form.freeShippingThreshold"
              :min="0"
              :step="1"
              style="width: 100%"
            />
            <small>購物車或直接下單金額達到此門檻時，運費會變成 0。</small>
          </el-form-item>
        </div>

        <div class="actions">
          <el-button type="primary" :loading="isSaving" @click="savePlatformConfig">
            {{ isSaving ? '儲存中...' : '儲存平台配置' }}
          </el-button>
        </div>
      </el-form>
    </el-card>
  </section>
</template>

<style scoped>
.platform-settings {
  display: grid;
  gap: 1rem;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 1rem;
}

.card-header .title {
  font-size: 1rem;
  font-weight: 700;
}

.subcopy {
  margin: 0.25rem 0 0;
  color: #909399;
  font-size: 13px;
}

:deep(.el-form-item small) {
  display: block;
  margin-top: 0.35rem;
  color: #909399;
  font-size: 12px;
  line-height: 1.5;
}

.shipping-grid {
  display: grid;
  gap: 1rem;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

.actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
}

@media (max-width: 900px) {
  .card-header {
    flex-direction: column;
    align-items: stretch;
  }

  .shipping-grid {
    grid-template-columns: 1fr;
  }
}
</style>
