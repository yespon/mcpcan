<template>
  <div>
    <el-dialog
      v-model="dialogInfo.visible"
      :title="'托管模式'"
      :show-close="false"
      :close-on-click-modal="false"
      class="access-type-dialog"
      width="620px"
      top="10vh"
    >
      <el-scrollbar height="70vh">
        <div class="px-5 py-2" v-loading="pageInfo.loading">
          <el-form
            ref="baseInfo"
            :model="pageInfo.formData"
            :rules="pageInfo.rules"
            label-width="auto"
            label-position="left"
          >
            <el-row :gutter="24">
              <el-col :span="18">
                <el-form-item :label="t('mcp.instance.formData.instanceName')" prop="name">
                  <el-input
                    v-model="pageInfo.formData.name"
                    :placeholder="t('mcp.instance.formData.instanceName')"
                  />
                </el-form-item>
                <el-form-item :label="t('mcp.template.formData.notes')" prop="notes">
                  <el-input
                    v-model="pageInfo.formData.notes"
                    :rows="2"
                    type="textarea"
                    :placeholder="t('mcp.template.formData.notes')"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <el-form-item prop="iconPath">
                  <Upload v-model="pageInfo.formData.iconPath"></Upload>
                </el-form-item>
              </el-col>
            </el-row>
            <el-form-item :label="t('mcp.instance.form.mcpProtocol')" prop="mcpProtocol">
              <el-segmented
                v-model="pageInfo.formData.mcpProtocol"
                :options="protocolOptions"
                @change="handleMcpProtocolChange"
              />
            </el-form-item>
            <el-row v-show="pageInfo.formData.mcpProtocol === 3">
              <el-col :span="18">
                <el-form-item prop="mcpServers">
                  <el-input
                    v-model="pageInfo.formData.mcpServers"
                    :rows="14"
                    type="textarea"
                    :placeholder="placeholderServer"
                    @blur="handleFormat"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="6">
                <div
                  class="pl-3 rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
                >
                  MCP服务SSE/STEAMABLE_HTTP协议配置当前为代理模式，流量会通过此平台网关转发到此配置提供的
                  MCP 配置中，保存后会在列表页显示网关访问配置。也可以查看
                  <a href="#/template-manage">模板列表</a> 提供了多个启动示例。
                </div>
              </el-col>
            </el-row>

            <el-form-item prop="packageId">
              <template #label>
                <span class="mr-2">{{ t('mcp.instance.formData.packageId') }}</span>
              </template>
              <div class="flex items-center">
                <span v-if="pageInfo.formData.packageId" class="mr-2 max-w-25 u-line-1">
                  <el-image :src="zipLogo" style="width: 16px; height: 16px"></el-image>
                  {{ currentPackage.name }}
                </span>
                <el-button size="small" class="base-btn" @click="selectVisible = true">
                  选择
                </el-button>
                <el-tooltip class="box-item" effect="dark" placement="top-start">
                  <el-button class="base-btn-link" link>下载示例代码包</el-button>
                  <template #content>
                    <div class="w-40">
                      示例 MCP
                      服务代码包，下载代码文件可用于了解不同编程语言在此平台的启动方式。也可以查看
                      模板列表 提供了多个代码包启动示例。
                    </div>
                  </template>
                </el-tooltip>
              </div>
            </el-form-item>

            <!-- 命令 -->
            <el-row>
              <el-col :span="18">
                <el-form-item :label="t('mcp.instance.formData.initScript')" prop="initScript">
                  <!-- <template #label>
                    <span class="mr-2">{{ t('mcp.instance.formData.initScript') }}</span>
                    <el-tooltip effect="light" placement="top" class="ml-6" :raw-content="true">
                      <el-icon><i class="icon iconfont MCP-tishi1"></i></el-icon>
                      <template #content>
                        <div class="w-36" v-html="t('mcp.instance.formData.initScriptTips')"></div>
                      </template>
                    </el-tooltip>
                  </template> -->
                  <el-input
                    v-model="pageInfo.formData.initScript"
                    :placeholder="t('mcp.instance.formData.initScript')"
                  />
                </el-form-item>
                <el-form-item :label="t('mcp.instance.formData.command')" prop="command">
                  <!-- <template #label>
                    <span class="mr-2">{{ t('mcp.instance.formData.command') }}</span>
                    <el-tooltip effect="light" placement="top" class="ml-6" :raw-content="true">
                      <el-icon><i class="icon iconfont MCP-tishi1"></i></el-icon>
                      <template #content>
                        <div v-html="pageInfo.tooltip.imgAddress"></div>
                      </template>
                    </el-tooltip>
                  </template> -->
                  <el-input
                    v-model="pageInfo.formData.command"
                    :placeholder="
                      showCommand ? InstanceData.COMMAND_TIP : t('mcp.instance.formData.command')
                    "
                    :disabled="showCommand"
                  />
                </el-form-item>
              </el-col>
              <el-col :span="6"></el-col>
            </el-row>
            <div class="font-size-3">
              <span class="font-bold">命令启动顺序：</span>
              <br />
              <span class="font-bold">代码包解压</span>
              :会自动下载代码文件到/app/codepkg/目录中并解压，注意在后续依赖命令和启动命令中先试用cd/app/codepkg后再执行。
              <br />
              <span>
                <span class="text-orange-400">特别注意:</span>
                假设压缩包code.zip中包含顶层文件夹code，解压后路径为/app/codepkg/code.
              </span>
              <br />
              <span class="font-bold">依赖命令</span>:当存在阻塞命名后会导致后续动作无法执行
              <br />
              <span class="font-bold">启动命令</span>: 默认在系统/根路径，如果上一步依赖命令中已经cd
              到项目目录，并且执行后没有退出，启动命令会沿用依赖中路径位置，不需要再次进入项目目录。
            </div>
            <div class="font-size-3 mt-4">
              <span class="font-bold light:text-black dark:text-white">启动容器说明</span>
              <div v-html="pageInfo.tooltip.imgAddress"></div>
            </div>

            <el-form-item label="访问地址" class="mt-6">
              <el-radio-group>
                <el-radio-button value="ip">
                  <el-popover
                    title="容器监听地址"
                    width="260"
                    trigger="click"
                    placement="bottom-start"
                  >
                    <template #reference>
                      <div class="w-24">0.0.0.0</div>
                    </template>
                    <div>
                      默认绑0.0.0.0地址，网关会自动识别容器并将流是导入容器中，当前运行模式无需修改此参数。
                    </div>
                  </el-popover>
                </el-radio-button>
                <el-radio-button value="port">
                  <el-popover
                    title="容器监听端口"
                    width="260"
                    trigger="click"
                    placement="bottom-start"
                  >
                    <template #reference>
                      <div class="w-24">8080</div>
                    </template>
                    <div>
                      默认8080，网关会自动识别容器并将流量导入容器中，当前运行模式无需修改 此参数。
                    </div>
                  </el-popover>
                </el-radio-button>
                <el-radio-button value="path">
                  <el-popover
                    title="容器监听地址"
                    width="260"
                    trigger="click"
                    placement="bottom-start"
                  >
                    <template #reference>
                      <div class="w-24">{{ '/see' || '/mcp' }}</div>
                    </template>
                    <div>
                      启动命令会将 STDIO 协议转为 SSE和 STEAMABLEHTTP协议，此路径 /sse
                      对应SSE协议，网关会自动识别路径并将流量导入容器中，当前运行模式无需修改此参数。

                      <!-- 启动命令会将 STIDO 协议转为 SSE和 STEAMABLEHTTP协议。此路径 /mcp 对应STEAMABLE
                      HTTP协议，网关会自动识别路径并将流量导入容器中，当前运行模式无需修 改此参数。 -->
                    </div>
                  </el-popover>
                </el-radio-button>
              </el-radio-group>
            </el-form-item>
            <!-- 环境默认 -->
            <el-form-item
              v-show="false"
              :label="t('mcp.instance.formData.environmentId')"
              prop="environmentId"
            >
              <el-select
                v-model="pageInfo.formData.environmentId"
                :placeholder="t('mcp.instance.formData.environmentId')"
                @change="handleChangeEnvironmentId"
                disabled
              >
                <el-option
                  v-for="(env, index) in envList"
                  :key="index"
                  :label="env.name"
                  :value="env.id"
                />
              </el-select>
            </el-form-item>
            <el-collapse :expand-icon-position="'left'">
              <el-collapse-item name="1">
                <template #title>
                  <div>
                    <span class="mr-1 font-bold">环境变量</span>
                    <span
                      class="rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
                    >
                      容器环境变量配置
                    </span>
                  </div>
                </template>
                <el-form-item prop="environmentVariables">
                  <!-- :label="t('mcp.instance.formData.environmentVariables')" -->
                  <div class="w-full mt-1 ml-1">
                    <div
                      class="flex align-center mb-2"
                      v-for="(env, index) in pageInfo.formData.environmentVariables"
                      :key="index"
                    >
                      <el-input
                        v-model="env.key"
                        :placeholder="t('mcp.instance.formData.key')"
                        class="mr-2"
                      />
                      <el-input
                        v-model="env.value"
                        :placeholder="t('mcp.instance.formData.value')"
                        class="mr-2"
                      />
                      <el-icon
                        class="cursor-pointer"
                        color="#F56C6C"
                        @click="handleDeleteEnvVariable(index)"
                      >
                        <Remove />
                      </el-icon>
                    </div>
                  </div>
                  <el-button class="add-env" :icon="Plus" plain @click="handleAddEnvVariable">
                    {{ t('mcp.instance.formData.add-env') }}
                  </el-button>
                </el-form-item>
                <div class="tip tip-primary">
                  注意:环境变量配置是指容器的环境变量配置，同时也会通过托管命令带入到MCP服务中
                </div>
              </el-collapse-item>
              <el-collapse-item v-if="pageInfo.formData.environmentId" name="2">
                <template #title>
                  <div>
                    <span class="mr-1 font-bold">卷挂载</span>
                    <span
                      class="rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
                    >
                      容器卷挂载配置
                    </span>
                  </div>
                </template>
                <el-form-item prop="volumeConfig">
                  <el-card
                    class="w-full mb-2"
                    v-for="(volume, index) in pageInfo.formData.volumeMounts"
                    :key="index"
                  >
                    <div>
                      <div class="flex align-center justify-between">
                        <span>
                          <span class="text-red">*</span>{{ t('mcp.instance.formData.volumeType') }}
                        </span>
                        <el-icon
                          color="#F56C6C"
                          class="cursor-pointer"
                          @click="handleDeleteVolume(index)"
                        >
                          <CircleClose />
                        </el-icon>
                      </div>
                      <el-select
                        v-model="volume.type"
                        :placeholder="t('mcp.instance.formData.volumeType')"
                        style="width: 200px"
                      >
                        <el-option label="HostPath" value="hostPath" />
                        <el-option
                          v-if="currentEnvironment.environment === 'kubernetes'"
                          label="PVC"
                          value="pvc"
                        />
                        <el-option
                          v-if="currentEnvironment.environment === 'docker'"
                          label="Volume"
                          value="volume"
                        />
                      </el-select>
                    </div>
                    <template v-if="currentEnvironment.environment === 'kubernetes'">
                      <template v-if="volume.type === 'hostPath'">
                        <div class="flex align-end">
                          <div>
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.nodeName') }}
                            </div>
                            <el-select
                              v-model="volume.nodeName"
                              :placeholder="t('mcp.instance.formData.nodeName')"
                              style="width: 200px"
                            >
                              <el-option
                                v-for="(node, index) in nodeList"
                                :key="index"
                                :label="node.name"
                                :value="node.name"
                              />
                            </el-select>
                          </div>
                          <el-button class="mr-2 ml-2" :icon="RefreshRight" />
                          <div>
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.hostPath') }}
                            </div>
                            <el-input
                              v-model="volume.hostPath"
                              :placeholder="t('mcp.instance.formData.hostPath')"
                            />
                          </div>
                        </div>
                        <div class="flex align-end">
                          <div class="mr-2">
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.mountPath') }}
                            </div>
                            <el-input
                              v-model="volume.mountPath"
                              :placeholder="t('mcp.instance.formData.mountPath')"
                              style="width: 200px"
                            />
                          </div>
                          <div>
                            <div>{{ t('mcp.instance.formData.readOnly') }}</div>
                            <el-switch v-model="volume.readOnly" />
                          </div>
                        </div>
                      </template>

                      <template v-if="volume.type === 'pvc'">
                        <div class="flex align-end">
                          <div>
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.pvcName') }}
                            </div>
                            <el-select
                              v-model="volume.pvcName"
                              placeholder="PVC"
                              style="width: 200px"
                              @change="handlePvcChange($event, volume)"
                            >
                              <el-option
                                v-for="(pvc, index) in pvcList"
                                :key="index"
                                :value="pvc.name"
                                :disabled="disabledPvcNode(pvc)"
                              >
                                <span>{{ pvc.name }}</span>
                                <el-tag v-if="disabledPvcNode(pvc)" color="orange">
                                  {{ t('mcp.template.formData.isBind') }}
                                </el-tag>
                              </el-option>
                            </el-select>
                          </div>
                          <el-button class="mr-2 ml-2" :icon="RefreshRight" />
                          <div class="mr-2">
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.mountPath') }}
                            </div>
                            <el-input
                              v-model="volume.mountPath"
                              :placeholder="t('mcp.instance.formData.mountPath')"
                            />
                          </div>
                          <div>
                            <div>{{ t('mcp.instance.formData.readOnly') }}</div>
                            <el-switch
                              v-model="volume.readOnly"
                              :disabled="disabledReadOnly(volume.pvcName)"
                            />
                          </div>
                        </div>
                        <el-space v-if="!selectedPvc(volume.pvcName)">
                          <el-text type="secondary">
                            {{ t('mcp.template.formData.accessModes') }}:
                            {{ selectedPvc(volume.pvcName).accessModes?.join(', ') }}
                          </el-text>
                          <template v-if="selectedPvc(volume.pvcName).pods?.length > 0">
                            <el-text type="secondary">
                              {{ t('mcp.template.formData.pods') }}:
                              {{ selectedPvc(volume.pvcName).pods.join(', ') }}
                            </el-text>
                          </template>
                        </el-space>
                      </template>
                    </template>
                    <template v-if="currentEnvironment.environment === 'docker'">
                      <template v-if="volume.type === 'hostPath'">
                        <div class="flex align-end">
                          <div class="mr-2">
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.hostPath') }}
                            </div>
                            <el-input
                              v-model="volume.hostPath"
                              :placeholder="t('mcp.instance.formData.hostPath')"
                            />
                          </div>
                          <div class="mr-2">
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.mountPath') }}
                            </div>
                            <el-input
                              v-model="volume.mountPath"
                              :placeholder="t('mcp.instance.formData.mountPath')"
                              style="width: 200px"
                            />
                          </div>
                          <div>
                            <div>{{ t('mcp.instance.formData.readOnly') }}</div>
                            <el-switch v-model="volume.readOnly" />
                          </div>
                        </div>
                      </template>

                      <template v-if="volume.type === 'volume'">
                        <div class="flex align-end">
                          <div>
                            <div><span class="text-red">*</span>{{ 'volumeName' }}</div>
                            <el-select
                              v-model="volume.volumeName"
                              placeholder="volumeName"
                              style="width: 200px"
                            >
                              <el-option
                                v-for="(pvc, index) in volumeList"
                                :key="index"
                                :value="pvc.name"
                              >
                                <span>{{ pvc.name }}</span>
                              </el-option>
                            </el-select>
                          </div>
                          <el-button class="mr-2 ml-2" :icon="RefreshRight" />
                          <div class="mr-2">
                            <div>
                              <span class="text-red">*</span
                              >{{ t('mcp.instance.formData.mountPath') }}
                            </div>
                            <el-input
                              v-model="volume.mountPath"
                              :placeholder="t('mcp.instance.formData.mountPath')"
                            />
                          </div>
                          <div>
                            <div>{{ t('mcp.instance.formData.readOnly') }}</div>
                            <el-switch v-model="volume.readOnly" />
                          </div>
                        </div>
                      </template>
                    </template>
                  </el-card>
                  <el-button class="add-env" :icon="Plus" plain @click="handleAddVolume">
                    {{ t('mcp.instance.formData.addVolume') }}
                  </el-button>
                </el-form-item>
              </el-collapse-item>
              <el-collapse-item name="3">
                <template #title>
                  <div>
                    <span class="mr-1 font-bold">Header 透传配置</span>
                    <span
                      class="rounded border border-[var(--ep-border-color-lighter)] text-[var(--ep-text-color-secondary)] text-xs leading-6 tracking-wide"
                    >
                      配置中存在 header 时，网关转发来自于客户端传输的 header时默认覆盖
                    </span>
                  </div>
                </template>
                <div>
                  <TokenForm ref="tokenForm" :formData="pageInfo.formData.tokens[0]"></TokenForm>
                </div>
                <div class="tip tip-primary">
                  注意:自定义header无需客户端提交，网关转发流量时自动携带到MCP服务请求中。当客户端请求，MCP配置，header透传自定义三者都存在header则优先级为header>MCP配置>客户端。也就是三者中存在相同header，以自定义配置为准，其次是MCP配置，再其次客户端请求haeder。
                </div>
              </el-collapse-item>
            </el-collapse>
          </el-form>
        </div>
      </el-scrollbar>
      <template #footer>
        <div class="text-center">
          <el-button @click="handleConfirm">保存并运行</el-button>
          <el-button @click="handleSaveAsTemplate">另存为模板</el-button>
          <el-button @click="handleClose">退出</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 代码包选择 -->
    <Select
      v-model="selectVisible"
      v-model:selected="pageInfo.formData.packageId"
      red="packageSelect"
      :title="t('mcp.instance.formData.packageId')"
      :options="packageList"
    >
      <template #options="{ option }">
        <div class="flex justify-between">
          <div class="flex align-center">
            <el-image :src="zipLogo" style="width: 32px; height: 32px"></el-image>
            <span class="ml-2"> {{ option.name }}</span>
          </div>
          <div class="flex align-center">
            {{ t('mcp.template.formData.size') }}：{{ formatFileSize(option.size) }}
            <span>{{ timestampToDate(option.createdAt) }}</span>
            <el-button
              type="primary"
              size="small"
              link
              class="base-btn-link ml-2"
              @click="handleViewCode(option)"
            >
              {{ t('mcp.instance.action.view') }}
            </el-button>
          </div>
        </div>
      </template>
    </Select>
  </div>
</template>

<script setup lang="ts">
import { Plus, Remove, CircleClose, RefreshRight, UploadFilled } from '@element-plus/icons-vue'
import { useInstanceFormHooks } from '../hooks/form-instance.ts'
import Upload from '@/components/upload/index.vue'
import { JsonFormatter } from '@/utils/json'
import TokenForm from './components/token-form.vue'
import zipLogo from '@/assets/logo/zip.png'
import { AccessType, McpProtocol, SourceType, InstanceData, NodeVisible } from '@/types/instance'
import { type VolumeMountsItme, type PvcForm, type Code } from '@/types/index.ts'
import { useMcpStoreHook } from '@/stores'
import Select from '@/components/mcp-select/index.vue'
import { formatFileSize, timestampToDate, getToken } from '@/utils/system'

const { t } = useI18n()
const {
  pageInfo,
  jumpToPage,
  placeholderServer,
  showCommand,
  disabledPvcNode,
  disabledReadOnly,
  selectedPvc,
  selectVisible,
} = useInstanceFormHooks()
const dialogInfo = ref({
  visible: false,
})
const { packageList, envList, nodeList, pvcList, volumeList, sourceOptions, accessTypeOptions } =
  toRefs(useMcpStoreHook())
const {
  handleGetPackageList,
  handleGetEnvList,
  handleGetNodeList,
  handleGetPvcList,
  handleGetVolumeList,
} = useMcpStoreHook()
const baseInfo = ref()
const protocolOptions = [
  { label: 'STDIO', value: 3 },
  { label: 'SSE', value: 1 },
  { label: 'STEAMABLE_HTTP', value: 2 },
]

/**
 * Current environment
 */
const currentEnvironment = computed(() => {
  return envList.value?.find((item: any) => item.id === pageInfo.value.formData.environmentId) || {}
})
/**
 * The current selected code package
 */
const currentPackage = computed(() => {
  return (
    packageList.value?.find(
      (item: { id: string }) => item.id === pageInfo.value.formData.packageId,
    ) || { name: '', size: '', createdAt: '' }
  )
})
const handleFormat = () => {
  pageInfo.value.formData.mcpServers = JsonFormatter.format(pageInfo.value.formData.mcpServers)
}
/**
 * Handle view code package
 */
const handleViewCode = (code: Code) => {
  jumpToPage({
    url: '/view-code-package',
    data: {
      id: code.id,
      name: code.name,
    },
    isOpen: true,
  })
}
/**
 * Handle add an environment variables
 */
const handleAddEnvVariable = () => {
  pageInfo.value.formData.environmentVariables.push({ key: '', value: '' })
}
/**
 * Handle delete environment variables
 * @param index - Index of environment variables
 */
const handleDeleteEnvVariable = (index: number | string) => {
  pageInfo.value.formData.environmentVariables.splice(Number(index), 1)
}
/**
 * Handle McpProtocol Changed
 */
const handleMcpProtocolChange = () => {}

/**
 * Handle delete Volume mounting
 * @param index - Index of need Delete Volume
 */
const handleDeleteVolume = (index: number | string) => {
  pageInfo.value.formData.volumeMounts.splice(Number(index), 1)
}
/**
 * Update readOnly status
 * @param key - $event
 * @param volume - item of pvc
 */
const handlePvcChange = (key: any, volume: VolumeMountsItme) => {
  const pvc = pvcList.value.find((pvc: PvcForm) => pvc.name === volume.pvcName)
  if (pvc) {
    const accessModes = pvc.accessModes || []
    let readOnlyValue = false
    if (accessModes.includes(NodeVisible.ROM)) {
      readOnlyValue = true
    } else if (accessModes.includes(NodeVisible.RWM)) {
      readOnlyValue = false
    } else if (accessModes.includes(NodeVisible.RWO)) {
      readOnlyValue = false
    }
    // Update readOnly status
    volume.readOnly = readOnlyValue
  }
}
/**
 * Handle add a volume
 */
const handleAddVolume = () => {
  pageInfo.value.formData.volumeMounts.push({
    type: '',
    nodeName: '', // HostPath name
    hostPath: '',
    mountPath: '',
    pvcName: '',
    volumeName: '',
    readOnly: false,
  })
}
/**
 * Handle change environment event
 */
const handleChangeEnvironmentId = async (e: number | undefined) => {
  pageInfo.value.formData.environmentId = e
  if (envList.value.find((item) => item.id === e)?.environment === 'docker') {
    handleGetVolumeList(pageInfo.value.formData.environmentId)
  } else {
    handleGetNodeList(pageInfo.value.formData.environmentId)
    handleGetPvcList(pageInfo.value.formData.environmentId)
  }
}
// Handle confirm save
const handleConfirm = () => {}
/**
 * save as a template
 */
const handleSaveAsTemplate = () => {}

const handleClose = () => {
  dialogInfo.value.visible = false
}

const init = async () => {
  dialogInfo.value.visible = true
  await handleGetEnvList() // 获取环境变量列表
  await handleGetPackageList() // 获取包列表

  nextTick(() => {
    pageInfo.value.formData.accessType = AccessType.HOSTING
    pageInfo.value.formData.mcpProtocol = 3
    // 默认选中第一个环境变量;所以上诉均没有return
    handleChangeEnvironmentId(envList.value[0]?.id)
  })
}

defineExpose({
  init,
})
</script>

<style lang="scss" scoped>
.add-env {
  width: 100%;
  border: 1px dashed var(--el-border-color);
}
.tip {
  padding: 10px;
  border-radius: 4px;
  font-size: 12px;
  &.tip-warning {
    background-color: #fff1f0;
    border-left: 5px solid var(--el-color-danger);
  }
  &.tip-primary {
    background-color: #409eff1a;
    border-left: 5px solid var(--el-color-primary);
  }
}
</style>
