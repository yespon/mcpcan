<template>
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
      <el-row v-if="pageInfo.formData.mcpProtocol === 3">
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
          <el-tooltip
            :disabled="!isPackageNameOverflow"
            :content="currentPackage.name"
            placement="top"
          >
            <div
              v-if="pageInfo.formData.packageId"
              class="mr-2 max-w-60 flex items-center"
              ref="packageNameRef"
            >
              <el-image
                :src="zipLogo"
                style="width: 16px; height: 16px"
                class="mr-1 shrink-0"
              ></el-image>
              <span class="u-line-1" @mouseenter="checkPackageNameOverflow">
                {{ currentPackage.name }}
              </span>
            </div>
          </el-tooltip>
          <mcp-button size="small" class="base-btn" @click="handleSelectPackage"> 选择 </mcp-button>
          <el-tooltip class="box-item" effect="light" placement="top-start">
            <el-button class="base-btn-link" link @click="downloadDialogVisible = true">
              下载示例代码包
            </el-button>
            <template #content>
              <div class="w-40">
                示例
                MCP服务代码包，下载代码文件可用于了解不同编程语言在此平台的启动方式。也可以查看<a
                  href="#/template-manage"
                  >模板列表</a
                >
                提供了多个代码包启动示例。
              </div>
            </template>
          </el-tooltip>
        </div>
      </el-form-item>

      <!-- 命令 -->
      <el-row :gutter="24">
        <el-col :span="16">
          <el-form-item :label="t('mcp.instance.formData.initScript')" prop="initScript">
            <el-input
              v-model="pageInfo.formData.initScript"
              :rows="6"
              type="textarea"
              :placeholder="t('mcp.instance.formData.initScript')"
            />
          </el-form-item>
          <el-form-item :label="t('mcp.instance.formData.command')" prop="command">
            <el-input
              v-model="pageInfo.formData.command"
              :rows="6"
              type="textarea"
              :placeholder="
                showCommand ? InstanceData.COMMAND_TIP : t('mcp.instance.formData.command')
              "
              :disabled="showCommand"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <div class="template-list-container">
            <div class="text-sm font-bold mb-2">运行环境命令实例</div>
            <el-popover
              :visible="!!selectedCommand.commands"
              placement="right"
              :title="selectedCommand?.name"
              :show-arrow="false"
              popper-style="width: 340px; padding: 0"
            >
              <template #reference>
                <div>
                  <el-scrollbar height="175px">
                    <div class="flex flex-col gap-2 pr-2 position-relative">
                      <div
                        v-for="(item, index) in deployTemplateData.configurations"
                        :key="index"
                        class="flex items-center gap-3 rounded-md mx-1 p-1 text-left type-item"
                        :class="{ 'active-type': selectedCommand?.name === item.name }"
                        @click="handleSelectCommand(item)"
                      >
                        <div class="font-bold truncate" :title="item.name">
                          {{ item.name }}
                          <!-- <el-tooltip
                                  effect="dark"
                                  :content="item.name"
                                  placement="top-start"
                                >
                                  {{ item.name }}
                                </el-tooltip> -->
                        </div>
                      </div>
                    </div>
                  </el-scrollbar>
                  <div class="tip tip-primary mt-2">
                    说明：平台针对
                    MCP服务启动提供了默认容器，选择以上运行环境示例可以查看启动示例命令。也可以到<a
                      href="#/template-manage"
                      >模板列表</a
                    >
                    中查看启动示例。
                  </div>
                </div>
              </template>
              <div v-if="selectedCommand" class="p-4 w-[300px]">
                <div class="mb-4">
                  <div class="flex justify-between items-center mb-1">
                    <span class="font-bold text-sm">依赖命令示例</span>
                    <el-button
                      link
                      type="primary"
                      @click="handleUseCommand(selectedCommand?.commands?.install, 'initScript')"
                      >使用</el-button
                    >
                  </div>
                  <div class="p-2 rounded text-xs text-gray-500 break-all">
                    {{ selectedCommand?.commands?.install }}
                  </div>
                </div>
                <div v-if="!showCommand" class="mb-2">
                  <div class="flex justify-between items-center mb-1">
                    <span class="font-bold text-sm">启动命令示例</span>
                    <el-button
                      link
                      type="primary"
                      @click="handleUseCommand(selectedCommand.commands.start, 'command')"
                      >使用</el-button
                    >
                  </div>
                  <div class="p-2 rounded text-xs text-gray-500 break-all">
                    {{ selectedCommand?.commands?.start }}
                  </div>
                </div>
                <div class="tip tip-primary mb-2">
                  {{ selectedCommand?.description || selectedCommand?.description_en }}
                </div>
                <div class="text-xs text-orange-500">注意：点击使用后会覆盖命令，无法恢复</div>
              </div>
            </el-popover>
          </div>
        </el-col>
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
            <el-popover title="容器监听地址" width="260" placement="bottom-start">
              <template #reference>
                <div class="w-24">0.0.0.0</div>
              </template>
              <div>
                默认绑0.0.0.0地址，网关会自动识别容器并将流是导入容器中，当前运行模式无需修改此参数。
              </div>
            </el-popover>
          </el-radio-button>
          <el-radio-button
            value="port"
            :class="pageInfo.formData.mcpProtocol !== 3 ? 'deep-form' : ''"
          >
            <el-popover title="容器监听端口" width="260" placement="bottom-start">
              <template #reference>
                <div v-if="pageInfo.formData.mcpProtocol === 3" class="w-24 flex items-center">
                  <div class="w-24">8080</div>
                </div>
                <div v-else class="w-30 flex items-center">
                  <el-input
                    v-model.number="pageInfo.formData.port"
                    type="number"
                    :placeholder="t('mcp.instance.formData.port')"
                    class="no-border-input w-24"
                  />
                </div>
              </template>
              <div>
                {{
                  pageInfo.formData.mcpProtocol === 3
                    ? '默认8080，网关会自动识别容器并将流量导入容器中，当前运行模式无需修改 此参数。'
                    : '请输入 MCP服务真实监听端口号'
                }}
              </div>
            </el-popover>
          </el-radio-button>
          <el-radio-button
            value="path"
            :class="pageInfo.formData.mcpProtocol !== 3 ? 'deep-form' : ''"
          >
            <el-popover title="容器监听路径" width="260" placement="bottom-start">
              <template #reference>
                <!-- STDIO 协议 -->
                <div
                  v-if="pageInfo.formData.mcpProtocol === 3"
                  class="w-24 flex items-center"
                  @click="handleChangePath"
                >
                  <div class="flex-1">{{ pageInfo.formData.servicePath }}</div>
                </div>
                <div v-else class="w-30 flex items-center">
                  <el-input
                    v-model="pageInfo.formData.servicePath"
                    :placeholder="t('mcp.template.formData.servicePath')"
                    class="no-border-input w-24"
                  />
                </div>
              </template>
              <div>
                {{
                  pageInfo.formData.mcpProtocol === 3
                    ? '启动命令会将 STDIO 协议转为 SSE和 STEAMABLEHTTP协议，此路径 /sse对应SSE协议，网关会自动识别路径并将流量导入容器中，当前运行模式无需修改此参数。'
                    : `请输入 MCP 服务挂载的访问 路径，留空则无前缀访问路 径。可参考 模板列表 中模版示例并启动后对照代码和容器日志。`
                }}
              </div>
            </el-popover>
          </el-radio-button>
        </el-radio-group>
        <el-radio-group class="mt-4">
          <el-radio-button value="ip">
            <el-popover title="容器监听地址" width="260" placement="bottom-start">
              <template #reference>
                <div class="w-24">0.0.0.0</div>
              </template>
              <div>
                默认绑0.0.0.0地址，网关会自动识别容器并将流是导入容器中，当前运行模式无需修改此参数。
              </div>
            </el-popover>
          </el-radio-button>
          <el-radio-button value="port">
            <el-popover title="容器监听端口" width="260" placement="bottom-start">
              <template #reference>
                <div class="w-24 flex items-center">
                  <div class="w-24">8080</div>
                </div>
              </template>
              <div>
                默认8080，网关会自动识别容器并将流量导入容器中，当前运行模式无需修改 此参数。
              </div>
            </el-popover>
          </el-radio-button>
          <el-radio-button value="path">
            <el-popover title="容器监听路径" width="260" placement="bottom-start">
              <template #reference>
                <div class="w-24 flex items-center" @click="handleChangePath">
                  <div class="flex-1">/mcp</div>
                </div>
              </template>
              <div>
                启动命令会将 STDIO 协议转为 SSE和 STEAMABLEHTTP协议，此路径
                /sse对应SSE协议，网关会自动识别路径并将流量导入容器中，当前运行模式无需修改此参数。
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.nodeName') }}
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.hostPath') }}
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.mountPath') }}
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.pvcName') }}
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.mountPath') }}
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.hostPath') }}
                      </div>
                      <el-input
                        v-model="volume.hostPath"
                        :placeholder="t('mcp.instance.formData.hostPath')"
                      />
                    </div>
                    <div class="mr-2">
                      <div>
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.mountPath') }}
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
                        <span class="text-red">*</span>{{ t('mcp.instance.formData.mountPath') }}
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
        <el-collapse-item v-if="!pageInfo.formData.instanceId" name="3">
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
          <span class="ml-2">{{ timestampToDate(option.createdAt) }}</span>
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
    <template #action>
      <el-upload
        class="mr-8"
        drag
        :action="action"
        :on-success="handleSuccess"
        :headers="headers"
        accept=".zip, .tar, .tar.gz, application/zip, application/x-tar, application/gzip"
        :auto-upload="true"
        :show-file-list="false"
      >
        <el-icon><UploadFilled /></el-icon>
        <div class="ml-2">
          {{ t('mcp.instance.openApi.localFile') }}
        </div>
      </el-upload>
    </template>
  </Select>

  <!-- 下载示例代码包弹窗 -->
  <el-dialog
    v-model="downloadDialogVisible"
    title="下载示例代码包"
    width="520px"
    append-to-body
    :close-on-click-modal="false"
  >
    <div class="flex flex-col gap-4">
      <div
        v-for="(item, index) in exampleList"
        :key="index"
        class="flex items-center justify-between p-4 border rounded-lg hover:shadow-md transition-shadow"
      >
        <div class="flex items-center gap-3">
          <el-image :src="zipLogo" style="width: 32px; height: 32px" />
          <div class="flex flex-col">
            <span class="font-bold text-base">{{ item.name }}</span>
            <span class="text-xs text-gray-500">{{ item.description }}</span>
          </div>
        </div>
        <el-button type="primary" link @click="handleDownloadExample(item)">
          <el-icon class="mr-1"><Download /></el-icon>
          下载
        </el-button>
      </div>
    </div>
  </el-dialog>
  <!-- probe instance dialog model -->
  <ProbeStatus ref="probe"></ProbeStatus>
  <ConfigDialog ref="config"></ConfigDialog>
  <LogDialog ref="log"></LogDialog>
</template>

<script setup lang="ts">
import {
  Plus,
  Remove,
  CircleClose,
  RefreshRight,
  Refresh,
  Download,
  UploadFilled,
} from '@element-plus/icons-vue'
import { useInstanceFormHooks } from '../../hooks/form-instance.ts'
import Upload from '@/components/upload/index.vue'
import McpButton from '@/components/mcp-button/index.vue'
import { JsonFormatter } from '@/utils/json'
import { ElMessage, ElMessageBox } from 'element-plus'
import TokenForm from './token-form.vue'
import zipLogo from '@/assets/logo/zip.png'
import { AccessType, McpProtocol, SourceType, InstanceData, NodeVisible } from '@/types/instance'
import {
  type VolumeMountsItme,
  type PvcForm,
  type Code,
  type InstanceResult,
} from '@/types/index.ts'
import { useMcpStoreHook } from '@/stores'
import Select from '@/components/mcp-select/index.vue'
import { formatFileSize, timestampToDate, getToken } from '@/utils/system'
import { InstanceAPI } from '@/api/mcp/instance'
import { TemplateAPI } from '@/api/mcp/template'
import { cloneDeep } from 'lodash-es'
import baseConfig from '@/config/base_config.ts'
import { Storage } from '@/utils/storage'
import ProbeStatus from '../probe-dialog.vue'
import ConfigDialog from '../config-dialog.vue'
import LogDialog from '../log-dialog.vue'
import deployTemplateData from '@/config/deploy-temlate-data.json'

const { t } = useI18n()
const {
  pageInfo,
  jumpToPage,
  originForm,
  placeholderServer,
  showCommand,
  disabledPvcNode,
  disabledReadOnly,
  selectedPvc,
  selectVisible,
} = useInstanceFormHooks()
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
const action = ref(
  baseConfig.SERVER_BASE_URL + (window as any).__APP_CONFIG__?.API_BASE + '/market/code/upload',
)
const headers = ref({
  Authorization: `Bearer ${Storage.get('token')}`,
})
const handleSuccess = (response: { code: number; data: { path: string } }) => {
  if (response.code !== 0) {
    return
  }
  handleGetPackageList()
  ElMessage.success(t('action.upload'))
}
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
const handleSelectPackage = () => {
  handleGetPackageList()
  selectVisible.value = true
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
const handleMcpProtocolChange = (servicePath: number) => {
  pageInfo.value.formData.servicePath = ['', '/sse', '/mcp', '/sse'][servicePath]
}

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

/**
 * Handle change path
 */
const handleChangePath = () => {
  pageInfo.value.formData.servicePath =
    pageInfo.value.formData.servicePath === '/sse' ? '/mcp' : '/sse'
}

const config = ref()
const handleConfig = () => {
  config.value.init(Object.assign(originForm.value, pageInfo.value.formData))
}
const probe = ref()
const handleViewStatus = () => {
  probe.value.init(pageInfo.value.formData)
}
const log = ref()
const handleViewLog = () => {
  log.value.init(pageInfo.value.formData)
}
// Handle confirm save
const handleConfirm = async () => {
  if (pageInfo.value.formData.instanceId) {
    const result = await ElMessageBox.confirm(
      t('mcp.instance.pageDesc.confirmEdit'),
      t('common.warn'),
      {
        confirmButtonText: t('common.ok'),
        cancelButtonText: t('common.cancel'),
        type: 'warning',
        customClass: 'tips-box',
        center: true,
        showClose: false,
        confirmButtonClass: 'is-plain el-button--danger danger-btn',
        customStyle: {
          width: '517px',
          height: '247px',
        },
      },
    )
    if (result !== 'confirm') return
  }
  baseInfo.value.validate(async (valid: boolean, fields: any) => {
    if (valid) {
      try {
        pageInfo.value.loading = true
        if (!pageInfo.value.formData.instanceId) {
          if (Array.isArray(pageInfo.value.formData.tokens[0].headers)) {
            pageInfo.value.formData.tokens[0].headers = Object.fromEntries(
              pageInfo.value.formData.tokens[0].headers?.map((header: any) => [
                header.key,
                header.value,
              ]),
            )
          }
        }
        const { instanceId } = await (
          pageInfo.value.formData.instanceId ? InstanceAPI.edit : InstanceAPI.create
        )({
          ...pageInfo.value.formData,
          environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
            (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
            {},
          ),
        })
        pageInfo.value.formData.instanceId = instanceId
        ElMessage.success(
          pageInfo.value.formData.instanceId ? t('action.edit') : t('action.create'),
        )
      } finally {
        pageInfo.value.loading = false
      }
    } else {
      ElMessage.warning(t('mcp.template.rules.validForm'))
    }
  })
}
/**
 * save as a template
 */
const handleSaveAsTemplate = () => {
  try {
    pageInfo.value.loading = true
    baseInfo.value.validate(async (valid: boolean) => {
      if (valid) {
        try {
          pageInfo.value.loading = true
          await TemplateAPI.create({
            ...pageInfo.value.formData,
            environmentVariables: pageInfo.value.formData.environmentVariables?.reduce(
              (obj: any, item: any) => ({ ...obj, [item.key]: item.value }),
              {},
            ),
          })
          ElMessage.success(t('action.create'))
        } finally {
          pageInfo.value.loading = false
        }
      }
    })
  } finally {
    pageInfo.value.loading = false
  }
}

/**
 * Handle get instance detail info
 */
const handleGetDetail = async (instance: InstanceResult) => {
  console.log(1111, instance)

  const data = await InstanceAPI.detail({
    instanceId: instance.instanceId,
  })
  pageInfo.value.formData = data
  pageInfo.value.accessType = data.accessType
  pageInfo.value.mcpServers = JsonFormatter.format(data.mcpServers)
  pageInfo.value.formData.environmentVariables = Object.keys(data.environmentVariables)?.map(
    (key) => ({ key, value: data.environmentVariables[key] }),
  )
  pageInfo.value.formData.volumeMounts = data.volumeMounts || []
}

const downloadDialogVisible = ref(false)
const exampleList = [
  {
    name: 'mcp-server-python',
    language: 'Python',
    description: 'Python 版本的 MCP Server 示例代码',
    url: '/static/code-package/mcp-example.zip',
  },
  {
    name: 'mcp-server-go',
    language: 'Go',
    description: 'Go 版本的 MCP Server 示例代码',
    url: '/static/code-package/mcp-example.zip',
  },
  {
    name: 'mcp-server-node',
    language: 'Node.js',
    description: 'Node.js 版本的 MCP Server 示例代码',
    url: '/static/code-package/mcp-example.zip',
  },
]

const handleDownloadExample = (item: any) => {
  // const link = document.createElement('a')
  // link.href = item.url
  // link.download = item.name + '.zip'
  // link.click()
  ElMessage.info('下载功能开发中')
}

const selectedCommand = ref({}) as any

const handleSelectCommand = (item: any) => {
  if (selectedCommand.value.name === item.name) {
    selectedCommand.value = {}
  } else {
    selectedCommand.value = item
  }
}

const handleUseCommand = (command: string, type: string) => {
  if (type === 'initScript') {
    pageInfo.value.formData.initScript = command
  } else if (type === 'command') {
    pageInfo.value.formData.command = command
  }
}
const packageNameRef = ref()
const isPackageNameOverflow = ref(false)
const checkPackageNameOverflow = () => {
  if (packageNameRef.value) {
    const el = packageNameRef.value
    isPackageNameOverflow.value = el.scrollWidth > el.clientWidth
  }
}

const init = async (instance: InstanceResult | null) => {
  await handleGetEnvList() // 获取环境变量列表
  await handleGetPackageList() // 获取包列表
  if (instance) {
    originForm.value = cloneDeep(instance)
    handleGetDetail(instance)
    return
  }
  nextTick(() => {
    pageInfo.value.formData.accessType = AccessType.HOSTING
    pageInfo.value.formData.mcpProtocol = 3
    pageInfo.value.formData.servicePath = '/sse'
    // 默认选中第一个环境变量;所以上诉均没有return
    handleChangeEnvironmentId(envList.value[0]?.id)
  })
}

defineExpose({
  init,
  handleConfig,
  handleViewStatus,
  handleViewLog,
  handleConfirm,
  handleSaveAsTemplate,
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
.upload-demo {
  width: 100%;
  color: var(--el-color-primary);
  :deep(.el-upload.is-drag) {
    height: 100%;
  }
  .el-icon--upload {
    color: var(--el-color-primary);
  }
  .el-upload__text {
    color: var(--el-color-primary);
  }
}
:deep(.el-input__wrapper) {
  background: var(--ep-bg-form);
}
:deep(.el-select__wrapper) {
  background: var(--ep-bg-form);
}
:deep(.el-textarea__inner) {
  background: var(--ep-bg-form);
}
:deep(.el-button) {
  background-color: transparent;
}
:deep(.deep-form .el-radio-button__inner) {
  padding: 0;
}
:deep(.no-border-input .el-input__wrapper) {
  box-shadow: none !important;
  // background: transparent !important;
  padding: 0 11px !important;
}
.type-item {
  transition: all 0.3s ease;
  overflow: hidden;
  background-color: var(--ep-bg-purple-color);
  border: 1px solid transparent;
  &.active-type {
    background-color: var(--ep-bg-purple-color-deep);
  }
  &:hover {
    scale: 1.02;
    cursor: pointer;
    background-color: var(--ep-bg-purple-color-deep);
    border-color: var(--ep-btn-color-top);
  }
}
</style>
